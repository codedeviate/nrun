package helper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func RunNPM(packageJSON PackageJSON, script string, args []string, envs map[string]string, flagList *FlagList, Version string) {
	if len(packageJSON.Scripts) > 0 {
		if len(packageJSON.Scripts[script]) > 0 {
			if len(packageJSON.Scripts["pre"+script]) > 0 {
				RunNPM(packageJSON, "pre"+script, args, envs, flagList, Version)
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

			if len(envs[script]) > 0 {
				if *flagList.BeVerbose {
					fmt.Println("============================================================")
					fmt.Println("Adding environment:", envs[script])
					if flagList.UsedPath != "" {
						fmt.Println("Using path:", flagList.UsedPath)
					}
					fmt.Println("============================================================")
				}
				cmd.Env = append(cmd.Environ(), envs[script])
			} else {
				if *flagList.BeVerbose {
					if flagList.UsedPath != "" {
						fmt.Println("============================================================")
						fmt.Println("Using path:", flagList.UsedPath)
						fmt.Println("============================================================")
					}
				}
				cmd.Env = append(cmd.Environ(), envs[script])
			}
			scriptNice := strings.Replace(script, ":", "_", -1)
			if len(envs[scriptNice]) > 0 {
				cmd.Env = append(cmd.Environ(), envs[scriptNice])
			}
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			runErr := cmd.Run()
			if runErr != nil {
				log.Println(runErr)
				return
			}
			if len(packageJSON.Scripts["post"+script]) > 0 {
				RunNPM(packageJSON, "post"+script, args, envs, flagList, Version)
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
