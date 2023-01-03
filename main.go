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
	"regexp"
	"sort"
	"strings"
)

const version = "0.8.7"

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Author          string            `json:"author"`
	License         string            `json:"license"`
	Dependencies    map[string]string `json:"dependencies"`
	Nyc             map[string]string `json:"nyc"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type Config struct {
	Env      map[string]map[string]string `json:"env"`
	Path     map[string]map[string]string `json:"path"`
	Projects map[string]string            `json:"projects"`
}

func ProcessPath(path string) (*PackageJSON, string, error) {
	if _, err := os.Stat(path + "/package.json"); errors.Is(err, os.ErrNotExist) {
		parts := strings.Split(path, "/")
		parts = parts[:len(parts)-1]
		path = strings.Join(parts, "/")
		if len(path) > 0 {
			return ProcessPath(path)
		}
		return nil, "", errors.New("no package.json found")
	}
	file, _ := os.ReadFile(path + "/package.json")
	packageJSON := PackageJSON{}
	_ = json.Unmarshal(file, &packageJSON)
	return &packageJSON, path, nil
}

func RunNPM(packageJSON PackageJSON, script string, args []string, envs map[string]string) {
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
			log.Printf("Can't find any script called \"%s\"\n", script)
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
	path, _ := cmd.Output()
	if len(path) > 0 {
		spath := strings.Trim(string(path), " \n")
		_, err := os.Stat(spath)
		return spath, err
	}
	return "", errors.New("can't find the requested shell")
}

func GetShell() (string, error) {
	envShell := os.Getenv("SHELL")
	if len(envShell) > 0 {
		return envShell, nil
	}
	// Try some magic to find shell
	if shell, err := GetShellByMagic("zsh"); !errors.Is(err, os.ErrNotExist) {
		return shell, nil
	}
	if shell, err := GetShellByMagic("bash"); !errors.Is(err, os.ErrNotExist) {
		return shell, nil
	}
	if shell, err := GetShellByMagic("sh"); !errors.Is(err, os.ErrNotExist) {
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

func main() {
	showScript := flag.Bool("s", false, "Show the script")
	showHelp := flag.Bool("h", false, "Show help")
	showList := flag.Bool("l", false, "Show all scripts")
	showVersion := flag.Bool("v", false, "Show current version")
	dummyCode := flag.Bool("d", false, "Exec some development dummy code")
	useAnotherPath := flag.String("p", "", "Use another path to find the package.json")
	flag.Parse()

	if *showHelp == true {
		fmt.Println("nrun - The npm script runner")
		fmt.Println("============================")
		fmt.Println("nrun will lookup the package.json used by the current project and execute the named script found in the scripts section of the package.json.")
		fmt.Println("")
		fmt.Println("Version: ", version)
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  nrun <script name> [args]  Run the script by name")
		fmt.Println("  nrun -l                    Shows all available scripts")
		fmt.Println("  nrun                       Shows all available scripts (same as the -l flag)")
		fmt.Println("  nrun -s <script name>      Show the script without running it")
		fmt.Println("  nrun -h                    Shows this help")
		fmt.Println("  nrun -v                    Shows current version")
		fmt.Println("")
		fmt.Println(".nrun.json in home directory")
		fmt.Println("===========================")
		fmt.Println("Often used script names can be mapped to other and shorter names in a file called .nrun.ini.")
		fmt.Println("This file should be placed in either the users home directory or in the same directory as the package.json.")
		fmt.Println("The format is more or less a standard ini-file. But there is one major difference. Section names can't contain colons and are therefor replaced with underscores.")
		fmt.Println("The section name is the full pathname of the directory that contains the package.json file.")
		fmt.Println("The section name must be a full path without any trailing slash.")
		fmt.Println("Environment variables can be defined by adding \"ENV:\" as a prefix to the sections name.")
		fmt.Println("These environment variables is not connected to the keys in the same directory but rather to the full script name.")
		fmt.Println("Global section names are \"*\" for mapping values and \"ENV:*\" for environment values. These values will be overridden by values defined in the specific directory.")
		fmt.Println("")
		fmt.Println("Example .nrun.json")
		fmt.Println("{")
		fmt.Println("  \"env\": {")
		fmt.Println("    \"/Users/codedeviate/Development/nruntest\": {")
		fmt.Println("      \"start:localhost\": \"PORT=3007\"")
		fmt.Println("    }")
		fmt.Println("  },")
		fmt.Println("  \"path\": {")
		fmt.Println("    \"/Users/codedeviate/Development/nruntest\": {")
		fmt.Println("      \"start\": \"start:localhost\",")
		fmt.Println("      \"test\": \"test:localhost:coverage\"")
		fmt.Println("    }")
		fmt.Println("  }")
		fmt.Println("}")
		fmt.Println("If you are in \"/Users/codedeviate/Development/nruntest\" and execute \"nrun start\" then that will be the same as executing \"PORT=3007 npm run start:localhost\" which is much shorter.")
		return
	}

	if *dummyCode == true {
		cmd := exec.Command("which", "bash")
		stdout, _ := cmd.Output()
		fmt.Println(strings.Trim(string(stdout), " \n"))
		cmd = exec.Command("which", "zsh")
		stdout, _ = cmd.Output()
		fmt.Println(strings.Trim(string(stdout), " \n"))
		cmd = exec.Command("which", "sh")
		stdout, _ = cmd.Output()
		fmt.Println(strings.Trim(string(stdout), " \n"))
		cmd = exec.Command("which", "ash")
		stdout, _ = cmd.Output()
		fmt.Println(strings.Trim(string(stdout), " \n"))
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
	if processErr != nil {
		log.Println(processErr)
	} else {
		if len(args) == 0 || *showList == true {
			ShowScripts(*packageJSON, defaultValues, defaultEnvironment)
		} else if *showScript == true {
			ShowScript(*packageJSON, script)
		} else {
			RunNPM(*packageJSON, script, args[1:], defaultEnvironment)
		}
	}
}
