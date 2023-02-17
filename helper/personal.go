package helper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
)

func ExecutePersonalFlags(flagList *FlagList) bool {
	if flagList.BeVerbose != nil && *flagList.BeVerbose {
		fmt.Println("###############################################")
		fmt.Println("Executing personal flags")
		fmt.Println("###############################################")
	}

	usr, _ := user.Current()
	homeDir := usr.HomeDir
	config, _ := ReadConfig(homeDir + "/.nrun.json")
	flagExecuted := false
	for i, i2 := range flagList.PersonalFlags {
		if i2 != nil && *i2 {
			if config.PersonalFlags[i] != nil && len(config.PersonalFlags[i]) > 0 {
				if flagList.BeVerbose != nil && *flagList.BeVerbose {
					fmt.Println("Running the personal flag", i)
				}
				shell, _ := GetShell()
				for _, command := range config.PersonalFlags[i] {
					cmd := exec.Command(shell, append([]string{"-c", command})...)

					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					runErr := cmd.Run()
					if runErr != nil {
						log.Println(runErr)
					}
					flagExecuted = true
				}
			}
		}
	}

	return flagExecuted
}
