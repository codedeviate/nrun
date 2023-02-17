package helper

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func ListProjectsFromConfig() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")
	if err != nil {
		return
	}
	maxLength := 0
	count := 0
	for k, _ := range config.Projects {
		count++
		if len(k) > maxLength {
			maxLength = len(k)
		}
	}

	if count > 0 {
		if count == 1 {
			fmt.Println("The following project is registered:")
		} else {
			fmt.Println("The following", count, "projects are registered:")
		}
	} else {
		fmt.Println("No projects are registered.")
	}
	for k, v := range config.Projects {
		fmt.Printf("%-*s : %s\n", maxLength, k, v)
	}
}

func AddProjectToConfig(args []string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")
	if err != nil {
		log.Println("Failed with", err)
		return
	}
	err = CopyFile(dir+"/.nrun.json", dir+"/.nrun.json.bak")
	if err != nil {
		log.Println("Failed with", err)
	} else {
		projPath := args[1]
		if len(projPath) > 1 && projPath[0:2] == ".." {
			cwd, _ := os.Getwd()
			projPath = cwd + "/" + projPath
		} else if projPath[0] == '.' {
			cwd, _ := os.Getwd()
			projPath = cwd + projPath[1:]
		}
		projPath, _ = filepath.Abs(projPath)
		if _, err := os.Stat(projPath); errors.Is(err, os.ErrNotExist) {
			log.Println("The path", "\""+projPath+"\"", "doesn't exists")
			return
		}
		if _, ok := config.Projects[args[0]]; ok {
			if config.Projects[args[0]] == projPath {
				log.Println("Project", "\""+args[0]+"\"", "already exists with this path")
				return
			}
			log.Println("Project", "\""+args[0]+"\"", "located at", "\""+config.Projects[args[0]]+"\"", "will be replaced with", "\""+projPath+"\"")
		}
		config.Projects[args[0]] = projPath
		err := WriteConfig(dir+"/.nrun.json", config)
		if err != nil {
			log.Println("Failed with", err)
		} else {
			log.Println("Project", "\""+args[0]+"\"", "added")
		}
	}
}

func RemoveProjectFromConfig(args []string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")
	err = CopyFile(dir+"/.nrun.json", dir+"/.nrun.json.bak")
	if err != nil {
		log.Println("Failed with", err)
	} else {
		delete(config.Projects, args[0])
		err := WriteConfig(dir+"/.nrun.json", config)
		if err != nil {
			log.Println("Failed with", err)
		}
		log.Println("Project", "\""+args[0]+"\"", "removed")
		if len(args) > 1 {
			args = args[1:]
			RemoveProjectFromConfig(args)
		}
	}
}

func GetProjectPath(args []string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")
	if err != nil {
		log.Println("Failed with", err)
		return
	} else {
		for _, arg := range args {
			if _, ok := config.Projects[arg]; ok {
				if len(args) == 1 {
					fmt.Println(config.Projects[arg])
				} else {
					fmt.Println(config.Projects[arg])
				}
			} else {
				log.Println("Project", "\""+args[0]+"\"", "doesn't exists")
			}
		}
	}
}
