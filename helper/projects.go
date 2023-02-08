package helper

import (
	"encoding/json"
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
	if _, err := os.Stat(dir + "/.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open(dir + "/.nrun.json")
		if err != nil {
			log.Println("Failed with", err)
		} else {
			byteValue, _ := os.ReadFile(jsonFile.Name())
			var config Config
			_ = json.Unmarshal(byteValue, &config)
			err := jsonFile.Close()
			if err != nil {
				return
			}
			maxLength := 0
			for k, _ := range config.Projects {
				if len(k) > maxLength {
					maxLength = len(k)
				}
			}

			if maxLength > 0 {
				fmt.Println("The following projects are registered:")
			} else {
				fmt.Println("No projects are registered.")
			}
			for k, v := range config.Projects {
				fmt.Printf("%-*s : %s\n", maxLength, k, v)
			}
		}
	}
}

func AddProjectToConfig(args []string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if _, err := os.Stat(dir + "/.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open(dir + "/.nrun.json")
		if err != nil {
			log.Println("Failed with", err)
		} else {
			byteValue, _ := os.ReadFile(jsonFile.Name())
			var config Config
			_ = json.Unmarshal(byteValue, &config)
			jsonFile.Close()
			err = CopyFile(jsonFile.Name(), jsonFile.Name()+".bak")
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
				jsonFile, err := os.Create(jsonFile.Name())
				if err != nil {
					log.Println("Failed with", err)
				} else {
					defer func(jsonFile *os.File) {
						err := jsonFile.Close()
						if err != nil {
							log.Println("Failed with", err)
						}
					}(jsonFile)
					jsonData, _ := json.MarshalIndent(config, "", "  ")
					_, err = jsonFile.Write(jsonData)
					if err != nil {
						log.Println("Failed with", err)
					}
					log.Println("Project", "\""+args[0]+"\"", "added")
				}
			}
		}
	}
}

func RemoveProjectFromConfig(args []string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if _, err := os.Stat(dir + "/.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open(dir + "/.nrun.json")
		if err != nil {
			log.Println("Failed with", err)
		} else {
			byteValue, _ := os.ReadFile(jsonFile.Name())
			var config Config
			_ = json.Unmarshal(byteValue, &config)
			err := jsonFile.Close()
			if err != nil {
				log.Println("Failed with", err)
				return
			}
			err = CopyFile(jsonFile.Name(), jsonFile.Name()+".bak")
			if err != nil {
				log.Println("Failed with", err)
			} else {
				delete(config.Projects, args[0])
				jsonFile, err := os.Create(jsonFile.Name())
				if err != nil {
					log.Println("Failed with", err)
				} else {
					defer func(jsonFile *os.File) {
						err := jsonFile.Close()
						if err != nil {
							log.Println("Failed with", err)
						}
					}(jsonFile)
					jsonData, _ := json.MarshalIndent(config, "", "  ")
					_, err = jsonFile.Write(jsonData)
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
		}
	}
}
