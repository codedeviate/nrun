package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const version = "0.12.0"

type PackageJSON struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	Main            string                 `json:"main"`
	Scripts         map[string]string      `json:"scripts"`
	Author          interface{}            `json:"author"`
	License         string                 `json:"license"`
	Dependencies    map[string]string      `json:"dependencies"`
	Nyc             map[string]interface{} `json:"nyc"`
	DevDependencies map[string]string      `json:"devDependencies"`
}

type Config struct {
	Env      map[string]map[string]string `json:"env"`
	Path     map[string]map[string]string `json:"path"`
	Projects map[string]string            `json:"projects"`
}

type LicenseList map[string][]string

func ProcessPath(path string, maxDepths ...int) (*PackageJSON, string, error) {
	var maxDepth int
	if len(maxDepths) == 0 {
		maxDepth = 100
	} else {
		maxDepth = maxDepths[0]
	}
	if _, err := os.Stat(path + "/package.json"); errors.Is(err, os.ErrNotExist) {
		parts := strings.Split(path, "/")
		parts = parts[:len(parts)-1]
		path = strings.Join(parts, "/")
		if maxDepth > 0 && len(path) > 0 {
			return ProcessPath(path, maxDepth-1)
		}
		return nil, "", errors.New("no package.json found")
	}
	file, _ := os.ReadFile(path + "/package.json")
	packageJSON := PackageJSON{}
	err := json.Unmarshal(file, &packageJSON)
	if err != nil {
		errorMessage := err.Error()
		return nil, "", errors.New("Error parsing package.json. " + strings.ToUpper(errorMessage[0:1]) + errorMessage[1:])
	}
	return &packageJSON, path, nil
}

func RunNPM(packageJSON PackageJSON, script string, args []string, envs map[string]string, beVerbose bool) {
	if len(packageJSON.Scripts) > 0 {
		if len(packageJSON.Scripts[script]) > 0 {
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
				if beVerbose {
					fmt.Println("====================")
					fmt.Println("Adding environment", envs[script])
					fmt.Println("====================")
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
		} else {
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
			if contains(validScripts, script) {
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
					fmt.Printf("nrun: {\n  nrun: '%s'\n},\nnpm: ", version)
				}
				runErr := cmd.Run()
				if runErr != nil {
					log.Println(runErr)
					return
				}
			} else {
				log.Println("Script", script, "does not exist")
			}
		}
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

func ShowScript(packageJSON PackageJSON, script string) {
	if len(packageJSON.Scripts) > 0 && len(packageJSON.Scripts[script]) > 0 {
		fmt.Printf("%s -> %s\n", script, packageJSON.Scripts[script])
	} else {
		log.Printf("Can't find any script called \"%s\"\n", script)
	}
}

func GetShellByMagic(key string) (string, error) {
	cmd := exec.Command("which", key)
	path, err := cmd.Output()

	if err == nil && len(path) > 0 {
		spath := strings.Trim(string(path), " \n")
		_, err := os.Stat(spath)
		return spath, err
	}
	return "", errors.New("shell not found")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func GetShell() (string, error) {
	envShell := os.Getenv("SHELL")
	if len(envShell) > 0 && FileExists(envShell) {
		return envShell, nil
	}
	// Try some magic to find shell
	if shell, err := GetShellByMagic("zsh"); err == nil {
		return shell, nil
	}
	if shell, err := GetShellByMagic("bash"); err == nil {
		return shell, nil
	}
	if shell, err := GetShellByMagic("sh"); err == nil {
		return shell, nil
	}
	return "", errors.New("can't find any shell")
}

func GetDefaultValues(path string) (map[string]string, map[string]string, map[string]string) {
	defaults := make(map[string]string, 1000)
	defaultEnvs := make(map[string]string, 1000)
	projects := make(map[string]string, 1000)
	usr, _ := user.Current()
	dir := usr.HomeDir
	if _, err := os.Stat(dir + "/.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open(dir + "/.nrun.json")
		if err != nil {
			log.Println("Failed with", err)
		} else {
			defer jsonFile.Close()
			byteValue, _ := os.ReadFile(jsonFile.Name())
			var config Config
			_ = json.Unmarshal(byteValue, &config)
			for k, v := range config.Path["*"] {
				defaults[k] = v
			}
			for k, v := range config.Path[path] {
				defaults[k] = v
			}
			for k, v := range config.Env["*"] {
				defaultEnvs[k] = v
			}
			for k, v := range config.Env[path] {
				defaultEnvs[k] = v
			}
			for k, v := range config.Projects {
				projects[k] = v
			}
		}
	}

	if path == "" {
		return defaults, defaultEnvs, projects
	}
	if _, err := os.Stat("./.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open("./.nrun.json")
		if err != nil {
			log.Println("Failed with", err)
		} else {
			defer jsonFile.Close()
			byteValue, _ := os.ReadFile(jsonFile.Name())
			var config Config
			_ = json.Unmarshal(byteValue, &config)
			for k, v := range config.Path["*"] {
				defaults[k] = v
			}
			for k, v := range config.Path[path] {
				defaults[k] = v
			}
			for k, v := range config.Env["*"] {
				defaultEnvs[k] = v
			}
			for k, v := range config.Env[path] {
				defaultEnvs[k] = v
			}
		}
	}

	return defaults, defaultEnvs, projects
}

func addProjectToConfig(args []string) {
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
			err = copyFile(jsonFile.Name(), jsonFile.Name()+".bak")
			if err != nil {
				log.Println("Failed with", err)
			} else {
				projPath := args[1]
				if projPath[0:2] == ".." {
					cwd, _ := os.Getwd()
					projPath = cwd + "/" + projPath
				} else if projPath[0] == '.' {
					cwd, _ := os.Getwd()
					projPath = cwd + projPath[1:]
				}
				projPath, _ = filepath.Abs(projPath)
				if _, err := os.Stat(projPath); errors.Is(err, os.ErrNotExist) {
					fmt.Println("The path", "\""+projPath+"\"", "doesn't exists")
					return
				}
				if _, ok := config.Projects[args[0]]; ok {
					if config.Projects[args[0]] == projPath {
						fmt.Println("Project", "\""+args[0]+"\"", "already exists with this path")
						return
					}
					fmt.Println("Project", "\""+args[0]+"\"", "located at", "\""+config.Projects[args[0]]+"\"", "will be replaced with", "\""+projPath+"\"")
				}
				config.Projects[args[0]] = projPath
				jsonFile, err := os.Create(jsonFile.Name())
				if err != nil {
					log.Println("Failed with", err)
				} else {
					defer jsonFile.Close()
					jsonData, _ := json.MarshalIndent(config, "", "  ")
					_, err = jsonFile.Write(jsonData)
					if err != nil {
						log.Println("Failed with", err)
					}
					fmt.Println("Project", "\""+args[0]+"\"", "added")
				}
			}
		}
	}
}

func removeProjectFromConfig(args []string) {
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
			err = copyFile(jsonFile.Name(), jsonFile.Name()+".bak")
			if err != nil {
				log.Println("Failed with", err)
			} else {
				delete(config.Projects, args[0])
				jsonFile, err := os.Create(jsonFile.Name())
				if err != nil {
					log.Println("Failed with", err)
				} else {
					defer jsonFile.Close()
					jsonData, _ := json.MarshalIndent(config, "", "  ")
					_, err = jsonFile.Write(jsonData)
					if err != nil {
						log.Println("Failed with", err)
					}
					fmt.Println("Project", "\""+args[0]+"\"", "removed")
					if len(args) > 1 {
						args = args[1:]
						removeProjectFromConfig(args)
					}
				}
			}
		}
	}
}

func listProjectsFromConfig() {
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

func copyFile(src string, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}
	return err
}

func showLicenseInfo(path string, licenseList LicenseList) LicenseList {
	if FileExists(path + "/package.json") {
		packageRaw, err := os.ReadFile(path + "/package.json")
		if err != nil {
			log.Println("Failed with", err)
		} else {
			packageJSON := PackageJSON{}
			err := json.Unmarshal(packageRaw, &packageJSON)
			if err != nil {
				log.Println("Failed opening package.json with", err)
			} else {
				// fmt.Printf("%s version %s, license %s\n", packageJSON.Name, packageJSON.Version, packageJSON.License)
				if packageJSON.License == "" {
					packageJSON.License = "UNKNOWN"
				}
				foundInLicense := false
				for _, name := range licenseList[packageJSON.License] {
					if name == packageJSON.Name {
						foundInLicense = true
						break
					}
				}
				if foundInLicense == false {
					if packageJSON.Name == "" {
						packageJSON.Name = path
					}
					licenseList[packageJSON.License] = append(licenseList[packageJSON.License], packageJSON.Name)
				}
			}
			if FileExists(path + "/node_modules") {
				files, err := os.ReadDir(path + "/node_modules")
				if err != nil {
					log.Println("Failed with", err)
				} else {
					for _, file := range files {
						if file.Name()[0] != '.' {
							if file.IsDir() {
								licenseList = showLicenseInfo(path+"/node_modules/"+file.Name(), licenseList)
							}
						}
					}
				}
			}
		}
	} else {
		files, err := os.ReadDir(path)
		if err != nil {
			log.Println("Failed with", err)
		} else {
			for _, file := range files {
				if file.Name() != "." && file.Name() != ".." {
					if file.IsDir() {
						licenseList = showLicenseInfo(path+"/"+file.Name(), licenseList)
					}
				}
			}
		}
	}
	return licenseList
}

func contains(stringsToSearch []string, key string) bool {
	for _, stringInList := range stringsToSearch {
		if strings.ToLower(stringInList) == key {
			return true
		}
	}
	return false
}

func wildMatch(stringsToSearch []string, key string) bool {
	for _, stringInList := range stringsToSearch {
		match, err := regexp.MatchString(".*"+strings.ToLower(stringInList)+".*", strings.ToLower(key))
		if match && err == nil {
			return true
		}
	}
	return false
}

func main() {
	showScript := flag.Bool("s", false, "Show the script")
	showHelp := flag.Bool("h", false, "Show help")
	showList := flag.Bool("l", false, "Show all scripts")
	showLicense := flag.Bool("L", false, "Show licenses of dependencies")
	showVersion := flag.Bool("v", false, "Show current version")
	dummyCode := flag.Bool("d", false, "Exec some development dummy code")
	useAnotherPath := flag.String("p", "", "Use another path to find the package.json")
	showCurrentProjectInfo := flag.Bool("i", false, "Show current project info")
	addProject := flag.Bool("ap", false, "Add a project to the config")
	removeProject := flag.Bool("rp", false, "Remove a project from the config")
	listProjects := flag.Bool("lp", false, "List all projects from the config")
	beVerbose := flag.Bool("V", false, "Be verbose, shows all environment variables set by nrun")

	flag.Parse()

	if *showHelp == true {
		fmt.Println("nrun - The npm script runner")
		fmt.Println("============================")
		fmt.Println("nrun will lookup the package.json used by the current project and execute the named script found in the scripts section of the package.json.")
		fmt.Println("")
		fmt.Println("Version: ", version)
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  nrun <script name> [args]         Run the script by name")
		fmt.Println("  nrun -i                           Show information about the current project")
		fmt.Println("  nrun -l                           Shows all available scripts")
		fmt.Println("  nrun                              Shows all available scripts (same as the -l flag)")
		fmt.Println("  nrun -s <script name>             Show the script without running it")
		fmt.Println("  nrun -h                           Shows this help")
		fmt.Println("  nrun -v                           Shows current version")
		fmt.Println("  nrun -lp                          List all projects from the config")
		fmt.Println("  nrun -ap <project name> <path>    Add a project to the config")
		fmt.Println("  nrun -rp <project name>           Remove a project from the config")
		fmt.Println("  nrun -L ([license name]) (names)  Shows all licenses of dependencies")
		fmt.Println("  nrun -V                           Shows all environment variables set by nrun")
		fmt.Println("")
		fmt.Println("For more information, see README.md")
		return
	}

	if *listProjects == true {
		listProjectsFromConfig()
		return
	}
	if *addProject == true {
		addProjectToConfig(flag.Args())
		return
	}
	if *removeProject == true {
		removeProjectFromConfig(flag.Args())
		return
	}
	if *dummyCode == true {
		fmt.Println("Dummy code that outputs the path to different shells if they are found")

		cmd := exec.Command("which", "sh")
		stdout, _ := cmd.Output()
		fmt.Println("Bourne Shell (sh)                 :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "bash")
		stdout, _ = cmd.Output()
		fmt.Println("GNU Bourne-Again Shell (bash)     :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "csh")
		stdout, _ = cmd.Output()
		fmt.Println("C Shell (csh)                     :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "ksh")
		stdout, _ = cmd.Output()
		fmt.Println("Korn Shell (ksh)                  :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "zsh")
		stdout, _ = cmd.Output()
		fmt.Println("Z Shell (zsh)                     :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "dash")
		stdout, _ = cmd.Output()
		fmt.Println("Debian Almquist Shell (dash)      :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "fish")
		stdout, _ = cmd.Output()
		fmt.Println("Friendly Interactive Shell (fish) :", strings.Trim(string(stdout), " \n"))

		cmd = exec.Command("which", "ash")
		stdout, _ = cmd.Output()
		fmt.Println("Almquist Shell (ash)              :", strings.Trim(string(stdout), " \n"))

		return
	}

	if *showVersion == true {
		fmt.Println(version)
		return
	}
	args := flag.Args()
	var script string
	if len(args) > 0 {
		script = args[0]
	}

	path, wdErr := os.Getwd()
	if wdErr != nil {
		log.Println(wdErr)
		return
	}
	if useAnotherPath == nil || *useAnotherPath == "" {
		env := os.Getenv("NRUNPROJECT")
		if env != "" {
			*useAnotherPath = env
		}
	}
	if useAnotherPath != nil && *useAnotherPath != "" {
		_, _, projects := GetDefaultValues("")
		path = *useAnotherPath
		if _, ok := projects[path]; ok {
			path = projects[path]
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			log.Println("The path is not a directory")
			return
		}
		os.Chdir(path)
	}
	packageJSON, path, processErr := ProcessPath(path)
	defaultValues, defaultEnvironment, _ := GetDefaultValues(path)

	if defaultValues != nil {
		if len(defaultValues[script]) > 0 {
			script = defaultValues[script]
		}
	}
	if *showLicense == true {
		licenseList := make(map[string][]string, 1000)
		licenseList = showLicenseInfo(path, licenseList)
		licenseListKeys := make([]string, 0, len(licenseList))
		for k := range licenseList {
			licenseListKeys = append(licenseListKeys, k)
		}
		sort.Strings(licenseListKeys)
		for index, key := range licenseListKeys {
			values := licenseList[licenseListKeys[index]]
			if len(args) == 0 || (contains(args, strings.ToLower(key)) || contains(args, "names") || wildMatch(args, key)) {
				if len(args) == 0 || (len(args) == 1 && contains(args, "names")) || (len(args) > 1 && contains(args, "names") && wildMatch(args, key)) || !contains(args, "names") {
					fmt.Println(key)
				}
				if !contains(args, "names") {
					licenseListValues := make([]string, 0, len(values))
					for k := range values {
						licenseListValues = append(licenseListValues, values[k])
					}
					sort.Strings(licenseListValues)
					for _, license := range licenseListValues {
						fmt.Println("  ", license)
					}
				}
			}
		}
		return
	}
	if *showCurrentProjectInfo == true {
		fmt.Println("Current project is", path)
		return
	}
	if processErr != nil {
		log.Println(processErr)
	} else {
		if len(args) == 0 || *showList == true {
			ShowScripts(*packageJSON, defaultValues, defaultEnvironment)
		} else if *showScript == true {
			ShowScript(*packageJSON, script)
		} else {
			RunNPM(*packageJSON, script, args[1:], defaultEnvironment, *beVerbose)
		}
	}
}
