package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"nrun/helper"
	"os"
	"os/user"
	"time"
)

const Version = "0.18.1"

func main() {
	originalPath, _ := os.Getwd()
	flagList := helper.ParseFlags()
	timeStarted := time.Now()

	//// Process any aliases first
	//config, _ := helper.ReadConfig(dir + "/.nrun.json")
	//aliasExecuted := false
	//for _, key := range os.Args[1:] {
	//	if config.Alias[key] != "" {
	//		helper.ExecuteAlias(config.Alias[key])
	//		aliasExecuted = true
	//	}
	//}
	//if aliasExecuted {
	//	return
	//}

	// Parse command line flags
	args := flag.Args()

	if flagList.UnpackJWTToken != nil && *flagList.UnpackJWTToken != "" {
		helper.UnpackJWTToken(*flagList.UnpackJWTToken)
		return
	}

	if flagList.SignJWTToken != nil && *flagList.SignJWTToken == true {
		helper.SignJWTToken(args)
		return
	}

	if flagList.PersonalFlags != nil && len(flagList.PersonalFlags) > 0 {
		if helper.ExecutePersonalFlags(flagList) {
			return
		}
	}

	if flagList.GetProjectPath != nil && *flagList.GetProjectPath {
		helper.GetProjectPath(args)
		return
	}

	var script string
	if len(args) > 0 {
		script = args[0]
		args = args[1:]
	}

	path, wdErr := os.Getwd()
	if wdErr != nil {
		log.Println(wdErr)
		return
	}

	defer func() {
		if flagList.MeasureTime != nil && *flagList.MeasureTime {
			duration := time.Since(timeStarted)
			timeElapsed := ""
			if int(duration.Minutes()) > 0 {
				timeElapsed = fmt.Sprintf("\nTime elapsed: %dmin %dsec\n", int(duration.Minutes()), int(duration.Seconds())-(int(duration.Minutes())*60))
			} else if int(duration.Seconds()) > 10 {
				timeElapsed = fmt.Sprintf("\nTime elapsed: %.1fsec\n", duration.Seconds())
			} else if int(duration.Seconds()) > 5 {
				timeElapsed = fmt.Sprintf("\nTime elapsed: %.2fsec\n", duration.Seconds())
			} else if int(duration.Seconds()) > 1 {
				timeElapsed = fmt.Sprintf("\nTime elapsed: %.3fsec\n", duration.Seconds())
			} else if int(duration.Milliseconds()) > 20 {
				timeElapsed = fmt.Sprintf("\nTime elapsed: %dms\n", int(duration.Milliseconds()))
			} else if int(duration.Microseconds()) > 20 {
				timeElapsed = fmt.Sprintf("\nTime elapsed: %d microseconds\n", int(duration.Microseconds()))
			} else {
				timeElapsed = fmt.Sprintf("\nTime elapsed:", duration)
			}
			fmt.Printf(timeElapsed)
			helper.Notify(timeElapsed)
		}
	}()

	if *flagList.SystemInfo {
		helper.SystemInfo(Version)
		return
	}

	if *flagList.ShowHelp == true {
		helper.ShowHelp(Version)
		return
	}

	if *flagList.ListProjects == true {
		helper.ListProjectsFromConfig()
		return
	}

	if *flagList.AddProject == true {
		helper.AddProjectToConfig(flag.Args())
		return
	}

	if *flagList.RemoveProject == true {
		helper.RemoveProjectFromConfig(flag.Args())
		return
	}

	if *flagList.DummyCode == true {
		helper.DummyCode()
		return
	}

	if *flagList.ShowVersion == true {
		fmt.Println(Version)
		return
	}

	if *flagList.UseAnotherPath == "" {
		env := os.Getenv("NRUNPROJECT")
		if env != "" {
			*flagList.UseAnotherPath = env
		}
	}

	if *flagList.UseAnotherPath != "" {
		_, _, projects, _, _ := helper.GetDefaultValues("")
		path = *flagList.UseAnotherPath
		if _, ok := projects[path]; ok {
			path = projects[path]
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			log.Println("The path is not a directory")
			return
		}
		os.Chdir(path)
	}
	packageJSON, path, processErr := helper.ProcessPath(path)
	defaultValues, defaultEnvironment, projects, scripts, vars := helper.GetDefaultValues(path)

	// Apply vars to all values from config
	defaultValues = helper.ApplyVars(defaultValues, vars)
	defaultEnvironment = helper.ApplyVars(defaultEnvironment, vars)
	projects = helper.ApplyVars(projects, vars)
	scripts = helper.ApplyVarsArray(scripts, vars)

	flagList.Vars = vars

	if flagList.ExecuteCommandInProjects != nil && *flagList.ExecuteCommandInProjects == true {
		helper.ExecuteCommandInProjects(path, script, args, defaultValues, defaultEnvironment, flagList, projects)
		return
	}

	if flagList.ExecuteCommand != nil && *flagList.ExecuteCommand == true {
		helper.ExecuteCommand(path, script, args, defaultValues, defaultEnvironment, flagList)
		return
	}

	if flagList.ExecuteMultipleScripts != nil && *flagList.ExecuteMultipleScripts == true {
		args = append([]string{script}, args...)
		helper.ExecuteMultipleScripts(args, flagList)
		return
	}

	if flagList.ExecuteScript != nil && *flagList.ExecuteScript == true {
		if len(scripts) > 0 && len(scripts[script]) > 0 {
			helper.ExecuteScripts(path, script, scripts[script], args, flagList)
		} else {
			log.Println("No script found")
		}
		return
	}

	if flagList.ExecuteScriptInProjects != nil && *flagList.ExecuteScriptInProjects == true {
		helper.ExecuteScriptList(script, scripts, args, projects, flagList)
		return
	}

	if flagList.ShowExecutableScript != nil && *flagList.ShowExecutableScript != "" {
		helper.ShowExecutableScript(scripts, flagList)
		return
	}

	if flagList.AddToExecutableScript != nil && *flagList.AddToExecutableScript != "" {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.AddToExecutableScript(args, flagList)
		return
	}

	if flagList.RemoveExecutableScript != nil && *flagList.RemoveExecutableScript != "" {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.RemoveExecutableScript(*flagList.RemoveExecutableScript, args)
		return
	}

	if flagList.ListExecutableScripts != nil && *flagList.ListExecutableScripts == true {
		helper.ListExecutableScripts(scripts, flagList)
		return
	}

	flagList.OriginalPath = originalPath
	flagList.UsedPath = path

	if processErr != nil {
		if packageJSON == nil {
			packageJSON = &helper.PackageJSON{}
		}
	}

	if defaultValues != nil {
		if len(defaultValues[script]) > 0 {
			script = defaultValues[script]
		}
	}

	if *flagList.ShowLicense == true {
		helper.ShowLicense(path, script, args)
		return
	}

	if *flagList.ShowCurrentProjectInfo == true {
		fmt.Println("Current project path is", path)
		return
	}

	if flagList.WebGetTemplate != nil && len(*flagList.WebGetTemplate) > 0 {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.WebGetTemplate(args, flagList)
		return
	}

	if flagList.WebGet != nil && *flagList.WebGet {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.WebGet(args, flagList)
		return
	}

	if flagList.ExecuteAlias != nil && *flagList.ExecuteAlias {
		usr, _ := user.Current()
		dir := usr.HomeDir
		config, _ := helper.ReadConfig(dir + "/.nrun.json")
		os.Chdir(flagList.UsedPath)
		for _, alias := range flag.Args() {
			command := config.Alias[alias]
			if len(command) > 0 {
				helper.ExecuteAlias(alias, command, flagList)
			}
		}
		return
	}

	if processErr != nil {
		if helper.PassthruNpm(*packageJSON, script, args, defaultEnvironment, Version) == false {
			log.Println(processErr)
		}
		return
	} else {
		if len(script) == 0 || *flagList.ShowList == true {
			helper.ShowScripts(*packageJSON, defaultValues, defaultEnvironment)
		} else if *flagList.ShowScript == true {
			helper.ShowScript(*packageJSON, script)
		} else {
			helper.RunNPM(*packageJSON, path, script, args, defaultEnvironment, flagList, Version)
		}
	}

}
