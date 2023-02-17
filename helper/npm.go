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

func RunNPM(packageJSON PackageJSON, path string, script string, args []string, envs map[string]string, flagList *FlagList, Version string) {
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
				RunNPM(packageJSON, path, "pre"+script, args, envs, flagList, Version)
			}
			runscript := packageJSON.Scripts[script]
			match, _err := regexp.Match(`^[^\s]*nrun(\s|$)`, []byte(runscript))
			if _err == nil && match {
				// This is a recursive call to nrun
				log.Println("Recursive call to nrun detected")
				return
			}
			args = append([]string{packageJSON.Scripts[script]}, args...)
			args = append([]string{"-c"}, args...)
			shell, shellErr := GetShell()
			if shellErr != nil {
				log.Println(shellErr)
				return
			}
			cmd := exec.Command(shell, args...)

			cmdEnv := cmd.Environ()

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

			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			runErr := cmd.Run()

			var exErr *exec.ExitError
			if errors.As(runErr, &exErr) {
				Notify("Process failed with error-code " + strconv.Itoa(exErr.ExitCode()))
				log.Println(runErr)
			} else if runErr != nil {
				log.Println(runErr)
				return
			}
			if len(packageJSON.Scripts["post"+script]) > 0 {
				RunNPM(packageJSON, path, "post"+script, args, envs, flagList, Version)
			}
		} else {
			if PassthruNpm(packageJSON, script, args, envs, Version) == false {
				log.Println("Script", script, "does not exist")
			}
		}
	} else {
		if PassthruNpm(packageJSON, script, args, envs, Version) == false {
			log.Println("No scripts defined in package.json")
		}
	}
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
