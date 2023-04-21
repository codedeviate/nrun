package helper

import (
	"errors"
	"fmt"
	"github.com/google/shlex"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

func RunNPM(packageJSON PackageJSON, path string, script string, args []string, envs map[string]string, flagList *FlagList, Version string, pipes map[string][]string) (int, error) {
	if flagList.BeVerbose != nil && *flagList.BeVerbose {
		fmt.Print("Running ", script, " in ", path, " with ")
		if len(args) > 0 {
			fmt.Println("args", strings.Join(args, ","))
		} else {
			fmt.Println("no args")
		}
	}
	if len(packageJSON.Scripts) > 0 {
		if len(packageJSON.Scripts[script]) > 0 {
			if len(packageJSON.Scripts["pre"+script]) > 0 {
				exitCode, err := RunNPM(packageJSON, path, "pre"+script, args, envs, flagList, Version, pipes)
				if err != nil {
					return exitCode, err
				}
			}
			runscript := packageJSON.Scripts[script]

			match, _err := regexp.Match(`^[^\s]*nrun(\s|$)`, []byte(runscript))
			if _err == nil && match {
				// This is a recursive call to nrun
				log.Println("Recursive call to nrun detected")
				return 0, nil
			}
			args = append([]string{runscript}, args...)
			args = append([]string{"-c"}, args...)
			shell, shellErr := GetShell()
			if shellErr != nil {
				log.Println(shellErr)
				return 0, shellErr
			}

			fmt.Println("Running", shell, strings.Join(args, " "))
			cmd := exec.Command(shell, args...)

			cmdEnv := cmd.Environ()

			// The difference between NoDefaultValues and NoDefaultValues2 is that NoDefaultValues2 removes the default values
			// from the config and NoDefaultValues only removes the default values from the current run
			if flagList.NoDefaultValues == nil || *flagList.NoDefaultValues == false {
				if len(envs[script]) > 0 {
					envParts, _ := shlex.Split(envs[script])
					for _, part := range envParts {
						cmdEnv = append(cmdEnv, part)
					}
					if *flagList.BeVerbose {
						fmt.Println("============================================================")
						fmt.Println("Adding environment:", envs[script])
						if flagList.UsedPath != "" {
							fmt.Println("Using path:", flagList.UsedPath)
						}
						fmt.Println("============================================================")
					}
				} else {
					if *flagList.BeVerbose {
						if flagList.UsedPath != "" {
							fmt.Println("============================================================")
							fmt.Println("Using path:", flagList.UsedPath)
							fmt.Println("============================================================")
						}
					}
				}
			}

			// Add node_modules/.bin to path if it exists
			node_bin_modules := path + "/node_modules/.bin"
			if IsDir(node_bin_modules) {
				cmdEnv = append(cmdEnv, "PATH="+node_bin_modules+":"+os.Getenv("PATH"))
			}

			// Add npm root -g to path if it exists (for global npm packages)
			npmRootGCmd := exec.Command("npm", "root -g")
			npmRootGCmdPathBytes, _ := npmRootGCmd.Output()
			npmRootGCmdPath := strings.Trim(string(npmRootGCmdPathBytes), " \n")
			if len(npmRootGCmdPath) > 0 && IsDir(npmRootGCmdPath) {
				cmdEnv = append(cmdEnv, "PATH="+npmRootGCmdPath+":"+os.Getenv("PATH"))
			}

			if *flagList.XAuthToken != "" {
				usr, _ := user.Current()
				dir := usr.HomeDir
				config, err := ReadConfig(dir + "/.nrun.json")
				if err == nil {
					if config.XAuthTokens[*flagList.XAuthToken] != "" {
						cmdEnv = append(cmdEnv, "X_AUTH_TOKEN="+config.XAuthTokens[*flagList.XAuthToken])
					} else {
						cmdEnv = append(cmdEnv, "X_AUTH_TOKEN="+*flagList.XAuthToken)
					}
				} else {
					cmdEnv = append(cmdEnv, "X_AUTH_TOKEN="+*flagList.XAuthToken)
				}
			}
			scriptNice := strings.Replace(script, ":", "_", -1)
			// Split this before we add it?
			if len(envs[scriptNice]) > 0 {
				cmdEnv = append(cmdEnv, envs[scriptNice])
			}

			// Manage overrides for env
			newEnv := []string{}
			overrideKeys := []string{}
			for _, envValue := range cmdEnv {
				if strings.HasPrefix(envValue, "OVERRIDE_") {
					newValue := strings.Replace(envValue, "OVERRIDE_", "", 1)
					overrideKeys = append(overrideKeys, strings.Split(newValue, "=")[0])
					newEnv = append(newEnv, newValue)
				}
			}

			finalEnv := []string{}
			for _, envValue := range cmdEnv {
				envKey := strings.Split(envValue, "=")[0]
				if !Contains(overrideKeys, envKey) {
					finalEnv = append(finalEnv, envValue)
				}
			}
			finalEnv = append(finalEnv, newEnv...)

			if flagList.BeVerbose != nil && *flagList.BeVerbose {
				if len(newEnv) > 0 {
					fmt.Println("============================================================")
					for _, override := range newEnv {
						fmt.Println("Overridden", override)
					}
					fmt.Println("============================================================")
				}
			}
			cmd.Env = finalEnv

			if flagList.ForcePipes != nil && *flagList.ForcePipes == true && len(pipes) > 0 && len(pipes[script]) > 0 {
				fmt.Println("Pipes are currently not supported")
				fmt.Println("They might be coming soon!")
				fmt.Println("Please use the flag -np to disable pipes for now")
				fmt.Println("Whatever happens from here on out is undefined behaviour")
				fmt.Println("The current status is: NOT WORKING AT ALL")

				fmt.Println("Pipes:", pipes[script])
				return 0, nil
			}
			if len(pipes) > 0 && len(pipes[script]) > 0 {
				if flagList.NoPipes == nil || *flagList.NoPipes == false {
					fmt.Println("Pipes are currently not supported")
					fmt.Println("They might be coming soon!")
					fmt.Println("Please use the flag -fp to test the usage of pipes....(if they are supported and what might happen)")
				}
			}

			//r, w, _ := os.Pipe()
			//cmd.Stdout = w
			//go func() {
			//	scanner := bufio.NewScanner(r)
			//	for scanner.Scan() {
			//		// fmt.Println(scanner.Text())
			//		fmt.Fprintf(os.Stdout, string(scanner.Bytes())+"\n")
			//	}
			//}()
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			runErr := cmd.Run()

			var exErr *exec.ExitError
			if errors.As(runErr, &exErr) {
				Notify("Process failed with error-code " + strconv.Itoa(exErr.ExitCode()))
				log.Println(runErr)
				return exErr.ExitCode(), runErr
			} else if runErr != nil {
				log.Println(runErr)
				return 0, runErr
			}

			if len(packageJSON.Scripts["post"+script]) > 0 {
				exitCode, err := RunNPM(packageJSON, path, "post"+script, args, envs, flagList, Version, pipes)
				if err != nil {
					return exitCode, err
				}
			}
		} else {
			if InternalCommands(packageJSON, script, args, envs, Version) == true {
				// Do nothing
			} else if PassthruNpm(packageJSON, script, args, envs, Version) == false {
				log.Println("Script", script, "does not exist")
			}
		}
	} else {
		if InternalCommands(packageJSON, script, args, envs, Version) == true {
			// Do nothing
		} else if PassthruNpm(packageJSON, script, args, envs, Version) == false {
			log.Println("No scripts defined in package.json")
		}
	}
	return 0, nil
}

func PassthruNpm(packageJSON PackageJSON, script string, args []string, envs map[string]string, Version string) bool {
	// Script names that are valid commands in npm
	validScripts := []string{
		"access",
		"adduser",
		"audit",
		"bin",
		"bugs",
		"cache",
		"ci",
		"completion",
		"config",
		"dedupe",
		"deprecate",
		"diff",
		"dist-tag",
		"docs",
		"doctor",
		"edit",
		"exec",
		"explain",
		"explore",
		"find-dupes",
		"fund",
		"get",
		"help",
		"hook",
		"init",
		"install",
		"install-ci-test",
		"install-test",
		"link",
		"ll",
		"login",
		"logout",
		"ls",
		"org",
		"outdated",
		"owner",
		"pack",
		"ping",
		"pkg",
		"prefix",
		"profile",
		"prune",
		"publish",
		"rebuild",
		"repo",
		"restart",
		"root",
		"run-script",
		"search",
		"set",
		"set-script",
		"shrinkwrap",
		"star",
		"stars",
		"start",
		"stop",
		"team",
		"test",
		"token",
		"uninstall",
		"unpublish",
		"unstar",
		"update",
		"version",
		"view",
		"whoami",
	}
	if len(script) == 0 || Contains(validScripts, script) {
		scriptArgs := []string{script}
		args = append(scriptArgs, args...)
		cmd := exec.Command("npm", args...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if script != "version" {
			fmt.Println("========================================")
			fmt.Println("Running \x1b[34mnpm", strings.Join(args[:], " "), "\x1b[0m")
			fmt.Println("========================================")
		} else {
			fmt.Printf("nrun: {\n  nrun: '%s'\n},\nnpm: ", Version)
		}
		runErr := cmd.Run()
		if runErr != nil {
			log.Println(runErr)
			return true
		}
		return true
	}
	return false
}
