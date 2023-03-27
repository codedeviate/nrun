package helper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ExecuteCommandInProjects(path string, script string, args []string, defaultValues map[string]string, defaultEnvironment map[string]string, flagList *FlagList, projects map[string]string, pipes map[string][]string) {
	if len(script) == 0 {
		log.Println("No command given 2")
		//			return
	}
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) > 1 && args[1] == "--" {
		tempArgs := args[0]
		args = args[2:]
		args = append([]string{tempArgs}, args...)
	}
	for projectName, projectPath := range projects {
		if flagList.BeVerbose != nil && *flagList.BeVerbose == true {
			fmt.Println("================================================================================")
			fmt.Println("Executing", script, strings.Join(args, " "))
			fmt.Println("  in project", projectName, "at", projectPath)
			fmt.Println("================================================================================")
		}
		ExecuteCommand(projectPath, script, args, defaultValues, defaultEnvironment, flagList, pipes)
		if flagList.BeVerbose != nil && *flagList.BeVerbose == true {
			fmt.Println("================================================================================")
		}
		fmt.Println("")
	}
}

func ExecuteCommand(path string, script string, args []string, defaultValues map[string]string, defaultEnvironment map[string]string, flagList *FlagList, pipes map[string][]string) {
	if len(script) == 0 {
		log.Println("No command given.")
		//		return
	}
	if len(args) > 0 && args[0] == "--" {
		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []string{}
		}
	}

	if flagList.BeVerbose != nil && *flagList.BeVerbose {
		fmt.Println("Executing command:", script, strings.Join(args, " "), "in", path)
	}
	os.Chdir(path)
	cmd := exec.Command(script, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	return
}
