package helper

import (
	"fmt"
	"github.com/google/shlex"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
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

			if flagList.NoDefaultValues == nil || *flagList.NoDefaultValues == false {
				if len(envs[script]) > 0 {
					envParts, _ := shlex.Split(envs[script])
					for _, part := range envParts {
						cmd.Env = append(cmd.Environ(), part)
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
				cmd.Env = append(cmd.Environ(), "PATH="+node_bin_modules+":"+os.Getenv("PATH"))
			}

			// Add npm root -g to path if it exists (for global npm packages)
			npmRootGCmd := exec.Command("npm", "root -g")
			npmRootGCmdPathBytes, _ := npmRootGCmd.Output()
			npmRootGCmdPath := strings.Trim(string(npmRootGCmdPathBytes), " \n")
			if len(npmRootGCmdPath) > 0 && IsDir(npmRootGCmdPath) {
				cmd.Env = append(cmd.Environ(), "PATH="+npmRootGCmdPath+":"+os.Getenv("PATH"))
			}

			if *flagList.XAuthToken != "" {
				usr, _ := user.Current()
				dir := usr.HomeDir
				config, err := ReadConfig(dir + "/.nrun.json")
				if err == nil {
					if config.XAuthTokens[*flagList.XAuthToken] != "" {
						cmd.Env = append(cmd.Environ(), "X_AUTH_TOKEN="+config.XAuthTokens[*flagList.XAuthToken])
					} else {
						cmd.Env = append(cmd.Environ(), "X_AUTH_TOKEN="+*flagList.XAuthToken)
					}
				} else {
					cmd.Env = append(cmd.Environ(), "X_AUTH_TOKEN="+*flagList.XAuthToken)
				}
			}
			scriptNice := strings.Replace(script, ":", "_", -1)
			if len(envs[scriptNice]) > 0 {
				cmd.Env = append(cmd.Environ(), envs[scriptNice])
			}

			// Manage overrides for env
			newEnv := []string{}
			overrides := []string{}
			overrideKeys := []string{}
			for _, envValue := range cmd.Env {
				if strings.HasPrefix(envValue, "OVERRIDE_") {
					newValue := strings.Replace(envValue, "OVERRIDE_", "", 1)
					overrideKeys = append(overrideKeys, strings.Split(newValue, "=")[0])
					overrides = append(overrides, newValue)
				} else {
					newEnv = append(newEnv, envValue)
				}
			}
			for _, overrideKey := range overrideKeys {
				for i := 0; i < len(newEnv); i++ {
					envValue := newEnv[i]
					if strings.HasPrefix(envValue, overrideKey+"=") {
						newEnv = append(newEnv[:i], newEnv[i+1:]...)
						i--
					}
				}
			}
			newEnv = append(newEnv, overrides...)
			cmd.Env = newEnv
			if flagList.BeVerbose != nil && *flagList.BeVerbose {
				if len(overrides) > 0 {
					fmt.Println("============================================================")
					for _, override := range overrides {
						fmt.Println("Overridden", override)
					}
					fmt.Println("============================================================")
				}
			}

			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			runErr := cmd.Run()
			for i, s := range cmd.Environ() {
				fmt.Println(i, s)
			}
			if runErr != nil {
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
