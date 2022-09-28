package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
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

type DefaultValues map[string]interface{}

func ProcessPath(path string) (*PackageJSON, string, error) {
	if _, err := os.Stat(path + "/package.json"); errors.Is(err, os.ErrNotExist) {
		parts := strings.Split(path, "/")
		parts = parts[:len(parts)-1]
		path = strings.Join(parts, "/")
		if len(path) > 0 {
			return ProcessPath(path)
		}
		return nil, "", errors.New("No package.json found")
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

func ShowScripts(packageJSON PackageJSON) {
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
	return "", errors.New("Can't find any shell")
}

func CollectDefaultValues(inifile string, path string) (map[string]string, error) {
	cfg, err := ini.Load(inifile)
	if err != nil {
		return nil, err
	}
	var defaults = make(map[string]string, 1000)
	if cfg.Section(path) != nil {
		keys := cfg.Section(path).Keys()
		for _, key := range keys {
			defaults[key.Name()] = key.Value()
		}
	}
	return defaults, nil
}

func GetDefaultValues(path string) (map[string]string, map[string]string) {
	defaults := make(map[string]string, 1000)
	defaultEnvs := make(map[string]string, 1000)
	usr, _ := user.Current()
	dir := usr.HomeDir
	if _, err := os.Stat(dir + "/.nrun.ini"); !errors.Is(err, os.ErrNotExist) {
		defWild, _ := CollectDefaultValues(dir+"/.nrun.ini", "*")
		if defWild != nil {
			for key, value := range defWild {
				defaults[key] = value
			}
		}
		defWildEnv, _ := CollectDefaultValues(dir+"/.nrun.ini", "ENV:*")
		if defWildEnv != nil {
			for key, value := range defWildEnv {
				defaultEnvs[key] = value
			}
		}
		def, _ := CollectDefaultValues(dir+"/.nrun.ini", path)
		if def != nil {
			for key, value := range def {
				defaults[key] = value
			}
		}
		defEnv, _ := CollectDefaultValues(dir+"/.nrun.ini", "ENV:"+path)
		if def != nil {
			for key, value := range defEnv {
				defaultEnvs[key] = value
			}
		}
	}
	if _, err := os.Stat("./.nrun.ini"); !errors.Is(err, os.ErrNotExist) {
		defWild, _ := CollectDefaultValues("./.nrun.ini", path)
		if defWild != nil {
			for key, value := range defWild {
				if len(defaults[key]) == 0 {
					defaults[key] = value
				}
			}
		}
		defWildEnv, _ := CollectDefaultValues("./.nrun.ini", "ENV:"+path)
		if defWildEnv != nil {
			for key, value := range defWildEnv {
				if len(defaultEnvs[key]) == 0 {
					defaultEnvs[key] = value
				}
			}
		}
		def, _ := CollectDefaultValues("./.nrun.ini", path)
		if def != nil {
			for key, value := range def {
				defaults[key] = value
			}
		}
		defEnv, _ := CollectDefaultValues("./.nrun.ini", "ENV:"+path)
		if def != nil {
			for key, value := range defEnv {
				defaultEnvs[key] = value
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
		fmt.Println("  nrun <scriptname> [args]   Run the script by name")
		fmt.Println("  nrun -l                    Shows all available scripts")
		fmt.Println("  nrun                       Shows all available scripts (same as the -l flag)")
		fmt.Println("  nrun -s <scriptname>       Show the script without running it")
		fmt.Println("  nrun -h                    Shows this help")
		fmt.Println("")
		fmt.Println(".nrun.ini in home directory")
		fmt.Println("===========================")
		fmt.Println("Often used scriptnames can be mapped to other and shorter names in a file called .nrun.ini.")
		fmt.Println("This file should be placed in either the users home directory or in the same directory as the package.json.")
		fmt.Println("The format is more or less a standard ini-file. But there is one major difference. Section names can't contain colons and are therefor replaced with underscores.")
		fmt.Println("The section name is the full pathname of the directoru that contains the package.json file.")
		fmt.Println("The section name must be a full path without any trailing slash.")
		fmt.Println("Environment variables can be defined by adding \"ENV:\" as a prefix to the sections name.")
		fmt.Println("These environment variables is not connected to the keys in the same directory but rather to the full script name.")
		fmt.Println("Global section names are \"*\" for mapping values and \"ENV:*\" for environment values. These values will be overridden by values defined in the specific directory.")
		fmt.Println("")
		fmt.Println("Example .nrun.ini")
		fmt.Println("[/Users/codedeviate/Development/nruntest]")
		fmt.Println("start=start:localhost")
		fmt.Println("[ENV:/Users/codedeviate/Development/nruntest]")
		fmt.Println("start_localhost=PORT=3007")
		fmt.Println("")
		fmt.Println("If you are in \"/Users/codedeviate/Development/nruntest\" and execute \"nrun start\" then that will be the same as executing \"PORT=3007 nrun start:localhost\" which is much shorter.")
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
		scriptNice := strings.Replace(script, ":", "_", -1)
		if len(defaultValues[scriptNice]) > 0 {
			script = defaultValues[scriptNice]
		}
	}
	if processErr != nil {
		fmt.Println(processErr)
	} else {
		if len(args) == 0 || *showList == true {
			ShowScripts(*packageJSON)
		} else if *showScript == true {
			ShowScript(*packageJSON, script)
		} else {
			RunNPM(*packageJSON, script, args[1:], defaultEnvironment)
		}
	}
}
