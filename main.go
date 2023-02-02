package main

import (
	"bufio"
	"bytes"
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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const version = "0.14.0"

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
	Scripts  map[string][]string          `json:"scripts"`
}

type LicenseList map[string][]string
type FlagList struct {
	showScript               *bool
	showHelp                 *bool
	showList                 *bool
	showLicense              *bool
	showVersion              *bool
	dummyCode                *bool
	useAnotherPath           *string
	usedPath                 string
	showCurrentProjectInfo   *bool
	addProject               *bool
	removeProject            *bool
	listProjects             *bool
	beVerbose                *bool
	passthruNpm              *bool
	systemInfo               *bool
	executeCommand           *bool
	executeCommandInProjects *bool
	executeScript            *bool
	listExecutableScripts    *bool
	addExecutableScript      *bool
	removeExecutableScript   *bool
	measureTime              *bool
}

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

func RunNPM(packageJSON PackageJSON, script string, args []string, envs map[string]string, flagList *FlagList) {
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
				if *flagList.beVerbose {
					fmt.Println("============================================================")
					fmt.Println("Adding environment:", envs[script])
					if flagList.usedPath != "" {
						fmt.Println("Using path:", flagList.usedPath)
					}
					fmt.Println("============================================================")
				}
				cmd.Env = append(cmd.Environ(), envs[script])
			} else {
				if *flagList.beVerbose {
					if flagList.usedPath != "" {
						fmt.Println("============================================================")
						fmt.Println("Using path:", flagList.usedPath)
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
		} else {
			if PassthruNpm(packageJSON, script, args, envs) == false {
				log.Println("Script", script, "does not exist")
			}
		}
	} else {
		if PassthruNpm(packageJSON, script, args, envs) == false {
			log.Println("No scripts defined in package.json")
		}
	}
}

func PassthruNpm(packageJSON PackageJSON, script string, args []string, envs map[string]string) bool {
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
	if len(script) == 0 || contains(validScripts, script) {
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
			return true
		}
		return true
	}
	return false
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

func GetDefaultValues(path string) (map[string]string, map[string]string, map[string]string, map[string][]string) {
	defaults := make(map[string]string, 1000)
	defaultEnvs := make(map[string]string, 1000)
	projects := make(map[string]string, 1000)
	scripts := make(map[string][]string, 1000)
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
			for k, v := range config.Scripts {
				scripts[k] = v
			}
		}
	}

	if path == "" {
		return defaults, defaultEnvs, projects, scripts
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

	return defaults, defaultEnvs, projects, scripts
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
				if len(projPath) > 1 && projPath[0:2] == ".." {
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

func GetVersionFromExecutable(executable string, args []string) string {
	cmd := exec.Command(executable, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

type Memory struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
}

func parseLine(raw string) (key string, value int) {
	fmt.Println(raw)
	text := strings.ReplaceAll(raw[:len(raw)-2], " ", "")
	keyValue := strings.Split(text, ":")
	intValue, _ := strconv.Atoi(keyValue[1])
	return keyValue[0], intValue * 1
}

func ReadMemoryStats() Memory {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return Memory{}
	}
	defer file.Close()
	bufio.NewScanner(file)
	scanner := bufio.NewScanner(file)
	res := Memory{}
	for scanner.Scan() {
		key, value := parseLine(scanner.Text())
		switch key {
		case "MemTotal":
			res.MemTotal = value
		case "MemFree":
			res.MemFree = value
		case "MemAvailable":
			res.MemAvailable = value
		}
	}
	return res
}

func SystemInfo() {
	fmt.Println("System information:")
	fmt.Println("  OS:", runtime.GOOS)
	if kernelVersion := GetVersionFromExecutable("uname", []string{"-r"}); version != "" {
		fmt.Println("  Kernel:", kernelVersion)
	}
	fmt.Println("  Architecture:", runtime.GOARCH)

	memory := ReadMemoryStats()
	if memory.MemTotal > 0 {
		fmt.Println("  Total memory:", memory.MemTotal/1024, "MB")
	}
	if memory.MemFree > 0 {
		fmt.Println("  Free memory:", memory.MemFree/1024, "MB")
	}
	if memory.MemAvailable > 0 {
		fmt.Println("  Available memory:", memory.MemAvailable/1024, "MB")
	}

	var version string
	fmt.Println("  Versions of installed tools:")
	if version = GetVersionFromExecutable("git", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`git\s+version\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Git:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("nrun", []string{"-v"}); version != "" {
		fmt.Println("    nrun:", version)
	}
	if version = GetVersionFromExecutable("node", []string{"-v"}); version != "" {
		fmt.Println("    Node:", version)
	}
	if version = GetVersionFromExecutable("npm", []string{"-v"}); version != "" {
		fmt.Println("    NPM:", version)
	}
	if version = GetVersionFromExecutable("go", []string{"version"}); version != "" {
		cleanVersion := regexp.MustCompile(`go\s+version\s+go([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Go:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("php", []string{"-v"}); version != "" {
		cleanVersion := regexp.MustCompile(`PHP\s+([0-9\.]+)\s+`).FindStringSubmatch(version)
		fmt.Println("    PHP:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("python3", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`Python\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Python3:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("ruby", []string{"-v"}); version != "" {
		cleanVersion := regexp.MustCompile(`ruby\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Ruby:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("gcc", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`clang\s+version\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    GCC:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("make", []string{"-v"}); version != "" {
		cleanVersion := regexp.MustCompile(`Make\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Make:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("ldd", []string{"--version"}); version != "" {
		fmt.Println("    LDD:", version)
	}
	if version = GetVersionFromExecutable("zig", []string{"version"}); version != "" {
		fmt.Println("    Zig:", version)
	}
	if version = GetVersionFromExecutable("bun", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`bun\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Bun:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("deno", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`deno\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Deno:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("rustc", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`rustc\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Rust:", cleanVersion[1])
	}
	if version = GetVersionFromExecutable("cargo", []string{"--version"}); version != "" {
		cleanVersion := regexp.MustCompile(`cargo\s+([0-9\.]+)`).FindStringSubmatch(version)
		fmt.Println("    Cargo:", cleanVersion[1])
	}
	fmt.Println("\nVersion information brought to you by nrun.")
}

func ExecuteCommand(path string, script string, args []string, defaultValues map[string]string, defaultEnvironment map[string]string, flagList *FlagList) {
	if len(script) == 0 {
		log.Println("No command given.")
		return
	}
	if len(args) > 0 && args[0] == "--" {
		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []string{}
		}
	}

	if flagList.beVerbose != nil && *flagList.beVerbose {
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

func ExecuteScripts(path string, scripts []string, args []string) {
	fmt.Println("================================================================================")
	fmt.Println("Warning: Executing script is currently an experimental work in progress.")
	fmt.Println("This feature might be removed in the future without prior notice.")
	fmt.Println("Please use at your own risk.")
	fmt.Println("Do not report issues about this feature. Unless you are willing to fix them.")
	fmt.Println("DO NOT USE THIS FEATURE IN PRODUCTION ENVIRONMENTS.")
	fmt.Println("================================================================================")
	fmt.Println("")
	if len(scripts) > 0 {
		os.Chdir(path)
		for _, script := range scripts {
			shell, shellErr := GetShell()
			if shellErr != nil {
				fmt.Println("Error:", shellErr)
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
}

func main() {
	originalWorkingDir, _ := os.Getwd()
	var flagList *FlagList
	flagList = new(FlagList)
	flagList.showScript = flag.Bool("s", false, "Show the script")
	flagList.showHelp = flag.Bool("h", false, "Show help")
	flagList.showList = flag.Bool("l", false, "Show all scripts")
	flagList.showLicense = flag.Bool("L", false, "Show licenses of dependencies")
	flagList.showVersion = flag.Bool("v", false, "Show current version")
	flagList.dummyCode = flag.Bool("d", false, "Exec some development dummy code")
	flagList.useAnotherPath = flag.String("p", "", "Use another path to find the package.json")
	flagList.showCurrentProjectInfo = flag.Bool("i", false, "Show current project info")
	flagList.addProject = flag.Bool("ap", false, "Add a project to the config")
	flagList.removeProject = flag.Bool("rp", false, "Remove a project from the config")
	flagList.listProjects = flag.Bool("lp", false, "List all projects from the config")
	flagList.beVerbose = flag.Bool("V", false, "Be verbose, shows all environment variables set by nrun")
	flagList.passthruNpm = flag.Bool("n", false, "Pass through to npm")
	flagList.systemInfo = flag.Bool("I", false, "Get the system information")
	flagList.executeCommand = flag.Bool("e", false, "Execute a command")
	flagList.executeCommandInProjects = flag.Bool("ep", false, "Execute a command in all projects")
	flagList.executeScript = flag.Bool("x", false, "Execute a script")
	flagList.listExecutableScripts = flag.Bool("lx", false, "List all executable scripts")
	// flagList.addExecutableScript = flag.Bool("ax", false, "Add an executable script")
	// flagList.removeExecutableScript = flag.Bool("rx", false, "Remove an executable script")
	flagList.measureTime = flag.Bool("T", false, "Measure the time it takes to execute the script")

	flag.Parse()

	timeStarted := time.Now()
	defer func() {
		if flagList.measureTime != nil && *flagList.measureTime {
			duration := time.Since(timeStarted)
			if int(duration.Minutes()) > 0 {
				fmt.Printf("\nTime elapsed: %dm %ds\n", int(duration.Minutes()), int(duration.Seconds())-(int(duration.Minutes())*60))
			} else if int(duration.Seconds()) > 10 {
				fmt.Printf("\nTime elapsed: %.1fs\n", duration.Seconds())
			} else if int(duration.Seconds()) > 5 {
				fmt.Printf("\nTime elapsed: %.2fs\n", duration.Seconds())
			} else if int(duration.Seconds()) > 1 {
				fmt.Printf("\nTime elapsed: %.3fs\n", duration.Seconds())
			} else if int(duration.Milliseconds()) > 20 {
				fmt.Printf("\nTime elapsed: %dms\n", int(duration.Milliseconds()))
			} else if int(duration.Microseconds()) > 20 {
				fmt.Printf("\nTime elapsed: %dus\n", int(duration.Microseconds()))
			} else {
				fmt.Println("\nTime elapsed:", duration)
			}
		}
	}()

	if *flagList.systemInfo {
		SystemInfo()
		return
	}
	if *flagList.showHelp == true {
		fmt.Println("nrun - The npm script runner")
		fmt.Println("============================")
		fmt.Println("nrun will lookup the package.json used by the current project and execute the named script found in the scripts section of the package.json.")
		fmt.Println("")
		fmt.Println("Version: ", version)
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  nrun <script name> [args]         Run the script by name")
		fmt.Println("  nrun -n                           Pass through to npm. Send everything to npm and let it handle it.")
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
		fmt.Println("  nrun -nv <node version>           Use a specific node version")
		fmt.Println("  nrun -e <command>                 Execute a command")
		fmt.Println("  nrun -ep <command>                Execute a command in all projects")
		fmt.Println("  nrun -T                           Measure the time it takes to execute the script")
		fmt.Println("For more information, see README.md")
		return
	}

	if *flagList.listProjects == true {
		listProjectsFromConfig()
		return
	}
	if *flagList.addProject == true {
		addProjectToConfig(flag.Args())
		return
	}
	if *flagList.removeProject == true {
		removeProjectFromConfig(flag.Args())
		return
	}
	if *flagList.dummyCode == true {
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

	if *flagList.showVersion == true {
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
	if *flagList.useAnotherPath == "" {
		env := os.Getenv("NRUNPROJECT")
		if env != "" {
			*flagList.useAnotherPath = env
		}
	}
	if *flagList.useAnotherPath != "" {
		_, _, projects, _ := GetDefaultValues("")
		path = *flagList.useAnotherPath
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
	defaultValues, defaultEnvironment, projects, scripts := GetDefaultValues(path)

	if flagList.executeCommandInProjects != nil && *flagList.executeCommandInProjects == true {
		if len(script) == 0 {
			log.Println("No command given")
			return
		}
		if len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		if len(args) > 1 && args[1] == "--" {
			tempArgs := args[0]
			args = args[2:]
			args = append([]string{tempArgs}, args...)
		}
		execArgs := args
		if len(execArgs) > 0 {
			execArgs = execArgs[1:]
		}
		for projectName, projectPath := range projects {
			if flagList.beVerbose != nil && *flagList.beVerbose == true {
				fmt.Println("================================================================================")
				fmt.Println("Executing", script, strings.Join(execArgs, " "))
				fmt.Println("  in project", projectName, "at", projectPath)
				fmt.Println("================================================================================")
			}
			ExecuteCommand(projectPath, script, execArgs, defaultValues, defaultEnvironment, flagList)
			if flagList.beVerbose != nil && *flagList.beVerbose == true {
				fmt.Println("================================================================================")
			}
			fmt.Println("")
		}
		return
	}

	if flagList.executeCommand != nil && *flagList.executeCommand == true {
		execArgs := args
		if len(execArgs) > 0 {
			execArgs = execArgs[1:]
		} else {
			execArgs = []string{}
		}
		ExecuteCommand(path, script, execArgs, defaultValues, defaultEnvironment, flagList)
		return
	}

	if flagList.executeScript != nil && *flagList.executeScript == true {
		if len(scripts) > 0 && len(scripts[script]) > 0 {
			ExecuteScripts(path, scripts[script], args[1:])
		}
		return
	}

	if flagList.listExecutableScripts != nil && *flagList.listExecutableScripts == true {
		fmt.Println("Executable scripts:")
		for scriptName, scriptScripts := range scripts {
			fmt.Println("  ", scriptName)
			for _, scriptScript := range scriptScripts {
				fmt.Println("    ->", scriptScript)
			}
		}
		return
	}

	if originalWorkingDir != path {
		flagList.usedPath = path
	}

	if processErr != nil {
		if packageJSON == nil {
			packageJSON = &PackageJSON{}
		}
		if len(args) > 0 {
			args = args[1:]
		}
		if PassthruNpm(*packageJSON, script, args, defaultEnvironment) == false {
			fmt.Println(processErr)
		}
		return
	}
	if defaultValues != nil {
		if len(defaultValues[script]) > 0 {
			script = defaultValues[script]
		}
	}
	if *flagList.showLicense == true {
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
	if *flagList.showCurrentProjectInfo == true {
		fmt.Println("Current project is", path)
		return
	}
	if processErr != nil {
		log.Println(processErr)
	} else {
		if len(args) == 0 || *flagList.showList == true {
			ShowScripts(*packageJSON, defaultValues, defaultEnvironment)
		} else if *flagList.showScript == true {
			ShowScript(*packageJSON, script)
		} else {
			RunNPM(*packageJSON, script, args[1:], defaultEnvironment, flagList)
		}
	}

}
