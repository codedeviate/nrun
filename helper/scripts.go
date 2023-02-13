package helper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecuteScriptList(script string, scripts map[string][]string, args []string, projects map[string]string, flagList *FlagList) {
	if len(scripts) > 0 && len(scripts[script]) > 0 {
		for projectName, projectPath := range projects {
			if flagList.BeVerbose != nil && *flagList.BeVerbose == true {
				fmt.Println("================================================================================")
				fmt.Println("Executing", script, strings.Join(args, " "))
				fmt.Println("  in project", projectName, "at", projectPath)
				fmt.Println("================================================================================")
			}
			ExecuteScripts(projectPath, script, scripts[script], args, flagList)
			if flagList.BeVerbose != nil && *flagList.BeVerbose == true {
				fmt.Println("================================================================================")
			}
			fmt.Println("")
		}
	} else {
		log.Println("No script found")
	}
}

func ScriptRunner(scripts []string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, script := range scripts {
		shell, shellErr := GetShell()
		if shellErr != nil {
			log.Println("Error:", shellErr)
			return
		}
		cmd := exec.Command(shell, append([]string{"-c", script})...)

		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		runErr := cmd.Run()
		if runErr != nil {
			log.Println(runErr)
			return
		}
	}
}

func ExecuteMultipleScripts(scripts []string, flagList *FlagList) {
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	config, _ := ReadConfig(homeDir + "/.nrun.json")
	ApplyVarsArray(config.Scripts, config.Vars)
	var wg sync.WaitGroup
	for _, script := range scripts {
		if flagList.BeVerbose != nil && *flagList.BeVerbose {
			fmt.Println("Executing script", script)
		}
		if len(config.Scripts[script]) > 0 {
			go ScriptRunner(config.Scripts[script], &wg)
			wg.Add(1)
		} else {
			log.Println("No script found for command", script)
		}
	}
	wg.Wait()
}

func ExecuteScripts(path string, scriptName string, scripts []string, args []string, flagList *FlagList) {
	if flagList.BeVerbose != nil && *flagList.BeVerbose {
		fmt.Println("Executing script", "\""+scriptName+"\"", "in", path)
	}
	if len(scripts) > 0 {
		os.Chdir(path)
		for _, script := range scripts {
			if flagList.BeVerbose != nil && *flagList.BeVerbose {
				fmt.Println("Executing command", "\""+script+"\"")
			}
			if len(script) > 2 {
				if script[0:2] == "@@" {
					doContinue := true
					script = script[2:]
					negate := false
					if len(script) > 0 && script[0] == '!' {
						negate = true
						script = script[1:]
					}
					if strings.Contains(script, ":") {
						commandParts := strings.Split(script, ":")
						if len(commandParts) > 1 {
							commandName := commandParts[0]
							commandArgs := strings.Join(commandParts[1:], ":")
							if commandName == "hasfile" || commandName == "hasfiles" {
								files := strings.Split(commandArgs, ",")
								fileFound := false
								for _, file := range files {
									file = strings.TrimSpace(file)
									if len(file) > 0 {
										if file[0] != '/' {
											file = path + "/" + file
										}
									}
									if FileExists(file) {
										fileFound = true
									} else if commandName == "hasfiles" {
										if negate {
											continue
										}
										return
									}
								}
								if !fileFound {
									if negate {
										continue
									}
									return
								} else if negate {
									return
								}
							} else if commandName == "cd" {
								commandArgs = strings.TrimSpace(commandArgs)
								if len(commandArgs) > 0 {
									if commandArgs[0] != '/' {
										os.Chdir(path + "/" + commandArgs)
										path, _ = os.Getwd()
									} else {
										os.Chdir(commandArgs)
										path, _ = os.Getwd()
									}
								}
							} else if commandName == "set" || commandName == "env" {
								commandArgs = strings.TrimSpace(commandArgs)
								if strings.Contains(commandArgs, "=") {
									commandParts := strings.Split(commandArgs, "=")
									if len(commandParts) > 1 {
										os.Setenv(commandParts[0], strings.Join(commandParts[1:], "="))
									}
								}
							} else if commandName == "unset" || commandName == "unenv" {
								commandArgs = strings.TrimSpace(commandArgs)
								os.Unsetenv(commandArgs)
							} else if commandName == "echo" {
								commandArgs = strings.TrimSpace(commandArgs)
								fmt.Println(commandArgs)
							} else if commandName == "isfile" {
								commandArgs = strings.TrimSpace(commandArgs)
								if len(commandArgs) > 0 {
									if commandArgs[0] != '/' {
										commandArgs = path + "/" + commandArgs
									}
									if !IsFile(commandArgs) {
										if negate {
											continue
										}
										return
									} else if negate {
										return
									}
								}
							} else if commandName == "isdir" {
								commandArgs = strings.TrimSpace(commandArgs)
								if len(commandArgs) > 0 {
									if commandArgs[0] != '/' {
										commandArgs = path + "/" + commandArgs
									}
									if !IsDir(commandArgs) {
										if negate {
											continue
										}
										return
									} else if negate {
										return
									}
								}
							}
						}
					} else {
						log.Println("Invalid command:", script)
						return
					}
					if doContinue {
						continue
					}
				}
			}
			shell, shellErr := GetShell()
			if shellErr != nil {
				log.Println("Error:", shellErr)
				return
			}
			cmd := exec.Command(shell, append([]string{"-c", script})...)

			env := os.Environ()
			env = append(env, []string{"NRUN_CURRENT_PATH=" + path}...)
			env = append(env, []string{"NRUN_CURRENT_SCRIPT=" + scriptName}...)
			env = append(env, []string{"NRUN_CURRENT_SCRIPT_CODE=" + script}...)
			for i, arg := range args {
				env = append(env, []string{"NRUN_ARG_" + strconv.Itoa(i) + "=" + arg}...)
			}
			cmd.Env = env

			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			runErr := cmd.Run()
			if runErr != nil {
				log.Println(runErr)
				return
			}
		}
	}
}

func ShowScript(packageJSON PackageJSON, script string) {
	if len(packageJSON.Scripts) > 0 && len(packageJSON.Scripts[script]) > 0 {
		fmt.Printf("%s -> %s\n", script, packageJSON.Scripts[script])
	} else {
		fmt.Printf("Can't find any script called \"%s\"\n", script)
	}
}

func ShowScripts(packageJSON PackageJSON, defaultValues map[string]string, defaultEnvironment map[string]string) {
	if len(packageJSON.Scripts) > 0 {
		fmt.Println("The following scripts are available")
		scripts := make([]string, 0, len(packageJSON.Scripts))
		for k := range packageJSON.Scripts {
			scripts = append(scripts, k)
		}
		sort.Strings(scripts)
		for _, s := range scripts {
			fmt.Printf(" - %s\n", s)
		}
	} else {
		log.Println("There are no scripts available")
	}
	if len(defaultValues) > 0 {
		fmt.Println("The following default values are available")
		for k, v := range defaultValues {
			fmt.Printf(" - %s: %s\n", k, v)
		}
	}
	if len(defaultEnvironment) > 0 {
		fmt.Println("The following default environment values are available")
		for k, v := range defaultEnvironment {
			fmt.Printf(" - %s: %s\n", k, v)
		}
	}
}

func ShowExecutableScript(scripts map[string][]string, flagList *FlagList) {
	fmt.Printf("Executable script %s:\n", *flagList.ShowExecutableScript)
	if len(scripts) > 0 && len(scripts[*flagList.ShowExecutableScript]) > 0 {
		for _, scriptScript := range scripts[*flagList.ShowExecutableScript] {
			fmt.Println("  ->", scriptScript)
		}
	} else {
		log.Println("  This script does not exist")
	}
	return
}

func ListExecutableScripts(scripts map[string][]string, flagList *FlagList) {
	fmt.Println("Executable scripts:")
	for scriptName, scriptScripts := range scripts {
		fmt.Println(" ", scriptName)
		for _, scriptScript := range scriptScripts {
			fmt.Println("    ->", scriptScript)
		}
	}
	return
}

func AddToExecutableScript(args []string, flagList *FlagList) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")
	err = CopyFile(dir+"/.nrun.json", dir+"/.nrun.json.bak")
	if err != nil {
		log.Println("Failed with", err)
	} else {
		var commands []string
		if config.Scripts != nil && config.Scripts[*flagList.AddToExecutableScript] != nil {
			commands = config.Scripts[*flagList.AddToExecutableScript]
		}
		commands = append(commands, strings.Join(args, " "))
		config.Scripts[*flagList.AddToExecutableScript] = commands
		err := WriteConfig(dir+"/.nrun.json", config)
		if err != nil {
			log.Println("Failed with", err)
			return
		}
		log.Println("Command added to the executable script", "\""+*flagList.AddToExecutableScript+"\"")
	}
}

func RemoveExecutableScript(script string, args []string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")

	err = CopyFile(dir+"/.nrun.json", dir+"/.nrun.json.bak")
	if err != nil {
		log.Println("Failed with", err)
	} else {
		if config.Scripts[script] == nil {
			log.Println("The script \"" + script + "\" doesn't exist")
			if len(args) > 0 {
				script = args[0]
				args = args[1:]
				RemoveExecutableScript(script, args)
			}
		} else {
			delete(config.Scripts, script)
			err := WriteConfig(dir+"/.nrun.json", config)
			if err != nil {
				log.Println("Failed with", err)
				return
			}
			log.Println("Executable script \"" + script + "\" has been removed")
			if len(args) > 0 {
				script = args[0]
				args = args[1:]
				RemoveExecutableScript(script, args)
			}
		}
	}
}
