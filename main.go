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

const Version = "0.21.0"

func main() {
	go helper.NotificationRunner()
	exitCode, err := process()

	for {
		time.Sleep(100 * time.Millisecond)
		if helper.WaitingNotifications == 0 {
			break
		}
	}

	if err != nil {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}

func process() (int, error) {
	originalPath, _ := os.Getwd()
	flagList := helper.ParseFlags()
	timeStarted := time.Now()

	// Parse command line flags
	args := flag.Args()

	if flagList.UnpackJWTToken != nil && *flagList.UnpackJWTToken != false {
		if len(args) == 0 {
			return 0, helper.UnpackJWTToken("")
		}
		for _, arg := range args {
			err := helper.UnpackJWTToken(arg)
			if err != nil {
				return 0, err
			}
		}
		return 0, nil
	}

	if flagList.SignJWTToken != nil && *flagList.SignJWTToken == true {
		return 0, helper.SignJWTToken(args)
	}
	if flagList.ValidateJWTToken != nil && *flagList.ValidateJWTToken != "" {
		return 0, helper.ValidateJWTToken(flagList, args)
	}

	if flagList.PersonalFlags != nil && len(flagList.PersonalFlags) > 0 {
		if helper.ExecutePersonalFlags(flagList) {
			return 0, nil
		}
	}

	if flagList.GetProjectPath != nil && *flagList.GetProjectPath {
		helper.GetProjectPath(args)
		return 0, nil
	}

	var script string
	if len(args) > 0 {
		script = args[0]
		args = args[1:]
	}

	path, wdErr := os.Getwd()
	if wdErr != nil {
		return 0, wdErr
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

	if flagList.Sleep != nil && *flagList.Sleep > 0 {
		time.Sleep(time.Duration(*flagList.Sleep) * time.Millisecond)
	}

	if flagList.TellAJoke != nil && *flagList.TellAJoke {
		helper.TellAJoke()
		return 0, nil
	}

	if *flagList.SystemInfo {
		helper.SystemInfo(Version)
		return 0, nil
	}

	if *flagList.ShowHelp == true {
		helper.ShowHelp(Version)
		return 0, nil
	}

	if *flagList.ListProjects == true {
		helper.ListProjectsFromConfig()
		return 0, nil
	}

	if *flagList.AddProject == true {
		helper.AddProjectToConfig(flag.Args())
		return 0, nil
	}

	if *flagList.RemoveProject == true {
		helper.RemoveProjectFromConfig(flag.Args())
		return 0, nil
	}

	if *flagList.DummyCode == true {
		helper.DummyCode()
		return 0, nil
	}

	if *flagList.ShowVersion == true {
		fmt.Println(Version)
		return 0, nil
	}

	if *flagList.VersionInformatrion == true {
		helper.VersionInformation()
		return 0, nil
	}
	if *flagList.UseAnotherPath == "" {
		env := os.Getenv("NRUNPROJECT")
		if env != "" {
			*flagList.UseAnotherPath = env
		}
	}

	if *flagList.UseAnotherPath != "" {
		_, _, projects, _, _, _, _ := helper.GetDefaultValues("")
		path = *flagList.UseAnotherPath
		if _, ok := projects[path]; ok {
			path = projects[path]
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			log.Println("The path is not a directory")
			return 0, nil
		}
		os.Chdir(path)
	}
	packageJSON, path, processErr := helper.ProcessPath(path)
	if processErr != nil {
		if packageJSON == nil {
			packageJSON = &helper.PackageJSON{}
		}
	}

	defaultValues, defaultEnvironment, projects, scripts, vars, packageJSONOverrides, pipes := helper.GetDefaultValues(path)

	// Check if we should skip overriding stuff
	if flagList.NoOverride != nil && *flagList.NoOverride == true {
		packageJSONOverrides = nil
		defaultValues = make(map[string]string)
		defaultEnvironment = make(map[string]string)
		scripts = make(map[string][]string)
	}

	// Check if we should skip overriding the package.json
	if flagList.NoPackageJSONOverride != nil && *flagList.NoPackageJSONOverride == true {
		packageJSONOverrides = nil
	}

	// Check if we should skip default values
	// The difference between NoDefaultValues and NoDefaultValues2 is that NoDefaultValues2 removes the default values
	// from the config and NoDefaultValues only removes the default values from the current run
	if flagList.NoDefaultValues2 != nil && *flagList.NoDefaultValues2 == true {
		defaultValues = make(map[string]string)
		defaultEnvironment = make(map[string]string)
	}

	// Check if we should override the package.json
	if packageJSONOverrides != nil {
		packageJSON = helper.ApplyPackageJSONOverrides(packageJSON, packageJSONOverrides)
	}

	// Apply vars to all values from config
	defaultValues = helper.ApplyVars(defaultValues, vars)
	defaultEnvironment = helper.ApplyVars(defaultEnvironment, vars)
	projects = helper.ApplyVars(projects, vars)
	scripts = helper.ApplyVarsArray(scripts, vars)
	pipes = helper.ApplyVarsArray(pipes, vars)

	flagList.Vars = vars

	if flagList.ExecuteCommandInProjects != nil && *flagList.ExecuteCommandInProjects == true {
		helper.ExecuteCommandInProjects(path, script, args, defaultValues, defaultEnvironment, flagList, projects, pipes)
		return 0, nil
	}

	if flagList.ExecuteCommand != nil && *flagList.ExecuteCommand == true {
		helper.ExecuteCommand(path, script, args, defaultValues, defaultEnvironment, flagList, pipes)
		return 0, nil
	}

	if flagList.ExecuteMultipleScripts != nil && *flagList.ExecuteMultipleScripts == true {
		args = append([]string{script}, args...)
		helper.ExecuteMultipleScripts(args, flagList)
		return 0, nil
	}

	if flagList.ExecuteScript != nil && *flagList.ExecuteScript == true {
		if len(scripts) > 0 && len(scripts[script]) > 0 {
			helper.ExecuteScripts(path, script, scripts[script], args, flagList)
		} else {
			log.Println("No script found")
		}
		return 0, nil
	}

	if flagList.ExecuteScriptInProjects != nil && *flagList.ExecuteScriptInProjects == true {
		helper.ExecuteScriptList(script, scripts, args, projects, flagList)
		return 0, nil
	}

	if flagList.ShowExecutableScript != nil && *flagList.ShowExecutableScript != "" {
		helper.ShowExecutableScript(scripts, flagList)
		return 0, nil
	}

	if flagList.AddToExecutableScript != nil && *flagList.AddToExecutableScript != "" {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.AddToExecutableScript(args, flagList)
		return 0, nil
	}

	if flagList.RemoveExecutableScript != nil && *flagList.RemoveExecutableScript != "" {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.RemoveExecutableScript(*flagList.RemoveExecutableScript, args)
		return 0, nil
	}

	if flagList.ListExecutableScripts != nil && *flagList.ListExecutableScripts == true {
		helper.ListExecutableScripts(scripts, flagList)
		return 0, nil
	}

	flagList.OriginalPath = originalPath
	flagList.UsedPath = path

	if defaultValues != nil {
		if len(defaultValues[script]) > 0 {
			script = defaultValues[script]
		}
	}

	if *flagList.ShowLicense == true {
		helper.ShowLicense(path, script, args)
		return 0, nil
	}

	if *flagList.ShowCurrentProjectInfo == true {
		fmt.Println("Current project path is", path)
		return 0, nil
	}

	if flagList.WebGetTemplate != nil && len(*flagList.WebGetTemplate) > 0 {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.WebGetTemplate(args, flagList)
		return 0, nil
	}

	if flagList.WebGet != nil && *flagList.WebGet {
		if len(script) > 0 {
			args = append([]string{script}, args...)
		}
		helper.WebGet(args, flagList)
		return 0, nil
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
		return 0, nil
	}

	//if processErr != nil {
	//	if helper.PassthruNpm(*packageJSON, script, args, defaultEnvironment, Version) == false {
	//		log.Println(processErr)
	//	}
	//	return
	//} else {
	if len(script) == 0 || *flagList.ShowList == true {
		helper.ShowScripts(*packageJSON, defaultValues, defaultEnvironment)
	} else if *flagList.ShowScript == true {
		helper.ShowScript(*packageJSON, script)
	} else {
		return helper.RunNPM(*packageJSON, path, script, args, defaultEnvironment, flagList, Version, pipes)
	}
	//}
	return 0, nil
}
