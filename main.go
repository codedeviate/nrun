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
	"sort"
	"strings"
)

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
	Env  map[string]map[string]string `json:"env"`
	Path map[string]map[string]string `json:"path"`
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
			args = append([]string{packageJSON.Scripts[script]}, args...)
			args = append([]string{"-c"}, args...)

			shell, shellErr := GetShell()
			if shellErr != nil {
				fmt.Println(shellErr)
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
				fmt.Println(runErr)
				return
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
		fmt.Println("There are no scripts available")
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
		fmt.Printf("Can't find any script called \"%s\"\n", script)
	}
}

func GetShell() (string, error) {
	envShell := os.Getenv("SHELL")
	if len(envShell) > 0 {
		return envShell, nil
	}
	if _, err := os.Stat("/bin/zsh"); !errors.Is(err, os.ErrNotExist) {
		return "/bin/zsh", nil
	}
	if _, err := os.Stat("/bin/bash"); !errors.Is(err, os.ErrNotExist) {
		return "/bin/bash", nil
	}
	return "", errors.New("can't find any shell")
}

func GetDefaultValues(path string) (map[string]string, map[string]string) {
	defaults := make(map[string]string, 1000)
	defaultEnvs := make(map[string]string, 1000)
	usr, _ := user.Current()
	dir := usr.HomeDir
	if _, err := os.Stat(dir + "/.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open(dir + "/.nrun.json")
		if err != nil {
			fmt.Println("Failed with", err)
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

	if _, err := os.Stat("./.nrun.json"); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open("./.nrun.json")
		if err != nil {
			fmt.Println("Failed with", err)
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

	return defaults, defaultEnvs
}

func main() {
	showScript := flag.Bool("s", false, "Show the script")
	showHelp := flag.Bool("h", false, "Show help")
	showList := flag.Bool("l", false, "Show all scripts")
	flag.Parse()

	if *showHelp == true {
		fmt.Println("nrun - The npm script runner")
		fmt.Println("============================")
		fmt.Println("nrun will lookup the package.json used by the current project and execute the named script found in the scripts section of the package.json.")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  nrun <script name> [args]  Run the script by name")
		fmt.Println("  nrun -l                    Shows all available scripts")
		fmt.Println("  nrun                       Shows all available scripts (same as the -l flag)")
		fmt.Println("  nrun -s <script name>      Show the script without running it")
		fmt.Println("  nrun -h                    Shows this help")
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
	packageJSON, path, processErr := ProcessPath(path)
	defaultValues, defaultEnvironment := GetDefaultValues(path)

	if defaultValues != nil {
		if len(defaultValues[script]) > 0 {
			script = defaultValues[script]
		}
	}
	if processErr != nil {
		fmt.Println(processErr)
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
