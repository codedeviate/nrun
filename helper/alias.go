package helper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func ExecuteAlias(alias string, command string, flagList *FlagList) {
	if flagList.BeVerbose != nil && *flagList.BeVerbose {
		fmt.Println("###############################################")
		fmt.Printf("Executing alias %s (%s)\n", alias, command)
		fmt.Println("###############################################")
	}
	shell, _ := GetShell()
	cmd := exec.Command(shell, append([]string{"-c", command})...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()
	if runErr != nil {
		log.Println(runErr)
		return
	}
}
