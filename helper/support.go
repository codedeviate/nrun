package helper

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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

func GetDefaultValues(path string) (map[string]string, map[string]string, map[string]string, map[string][]string, map[string]string, map[string]interface{}, map[string][]string) {
	defaults := make(map[string]string, 1000)
	defaultEnvs := make(map[string]string, 1000)
	projects := make(map[string]string, 1000)
	scripts := make(map[string][]string, 1000)
	vars := make(map[string]string, 1000)
	packageJson := make(map[string]interface{}, 1000)
	pipes := make(map[string][]string, 1000)

	usr, _ := user.Current()
	dir := usr.HomeDir
	config, err := ReadConfig(dir + "/.nrun.json")
	if err != nil {
		if err.Error() != "config file not found" {
			log.Println("Failed global .nrun.json with", err)
		}
	} else {
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
		for k, v := range config.Vars {
			vars[k] = v
		}
		for k, v := range config.Projects {
			projects[k] = v
		}
		for k, v := range config.Scripts {
			scripts[k] = v
		}
		for k, v := range config.Pipes {
			klist := strings.Split(k, ",")
			for _, k2 := range klist {
				k2 = strings.TrimSpace(k)
				if k2[0] == '@' && len(config.Projects[k2[1:]]) > 0 && config.Projects[k2[1:]] == path {
					for k3, v2 := range v {
						pipes[k3] = v2
					}
				} else if k2 == "*" {
					for k3, v2 := range v {
						pipes[k3] = v2
					}
				} else if k2 == path {
					for k3, v2 := range v {
						pipes[k3] = v2
					}
				}
			}
		}
		if config.PackageJSONOverride != nil {
			if packageJson["scripts"] == nil {
				packageJson["scripts"] = make(map[string]interface{}, 1000)
			}
			if config.PackageJSONOverride["*"] != nil {
				for k, v := range config.PackageJSONOverride["*"].(map[string]interface{}) {
					vType := reflect.TypeOf(v)
					if vType.Kind() == reflect.Map {
						for k2, v2 := range v.(map[string]interface{}) {
							if k == "scripts" {
								packageJson["scripts"].(map[string]interface{})[k2] = v2
							}
						}
					}
				}
			}
			for kList, v := range config.PackageJSONOverride {
				kList := strings.Split(kList, ",")
				for _, k := range kList {
					k = strings.TrimSpace(k)
					if path == k {
						vType := reflect.TypeOf(v)
						if vType.Kind() == reflect.Map {
							for k2, v2 := range v.(map[string]interface{}) {
								for k3, v3 := range v2.(map[string]interface{}) {
									if k2 == "scripts" {
										packageJson["scripts"].(map[string]interface{})[k3] = v3
									}
								}
							}
						}
					}
					if len(k) > 1 && k[0] == '@' && len(config.Projects[k[1:]]) > 0 && config.Projects[k[1:]] == path {
						vType := reflect.TypeOf(v)
						if vType.Kind() == reflect.Map {
							for k2, v2 := range v.(map[string]interface{}) {
								for k3, v3 := range v2.(map[string]interface{}) {
									if k2 == "scripts" {
										packageJson["scripts"].(map[string]interface{})[k3] = v3
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if path == "" {
		return defaults, defaultEnvs, projects, scripts, vars, packageJson, pipes
	}
	config, err = ReadConfig("./.nrun.json")
	if err != nil {
		if err.Error() != "config file not found" {
			log.Println("Failed local .nrun.json with", err)
		}
	} else {
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
		for k, v := range config.Vars {
			vars[k] = v
		}
		for k, v := range config.Pipes {
			klist := strings.Split(k, ",")
			for _, k2 := range klist {
				k2 = strings.TrimSpace(k)
				if k2[0] == '@' && len(config.Projects[k2[1:]]) > 0 && config.Projects[k2[1:]] == path {
					for k3, v2 := range v {
						pipes[k3] = v2
					}
				} else if k2 == "*" {
					for k3, v2 := range v {
						pipes[k3] = v2
					}
				} else if k2 == path {
					for k3, v2 := range v {
						pipes[k3] = v2
					}
				}
			}
		}
		if config.PackageJSONOverride != nil {
			if packageJson["scripts"] == nil {
				packageJson["scripts"] = make(map[string]interface{}, 1000)
			}
			for kList, v := range config.PackageJSONOverride {
				kList := strings.Split(kList, ",")
				for _, k := range kList {
					k = strings.TrimSpace(k)
					if path == k {
						vType := reflect.TypeOf(v)
						if vType.Kind() == reflect.Map {
							for k2, v2 := range v.(map[string]interface{}) {
								for k3, v3 := range v2.(map[string]interface{}) {
									if k2 == "scripts" {
										packageJson["scripts"].(map[string]interface{})[k3] = v3
									}
								}
							}
						}
					}
					if len(k) > 1 && k[0] == '@' && len(config.Projects[k[1:]]) > 0 && config.Projects[k[1:]] == path {
						vType := reflect.TypeOf(v)
						if vType.Kind() == reflect.Map {
							for k2, v2 := range v.(map[string]interface{}) {
								for k3, v3 := range v2.(map[string]interface{}) {
									if k2 == "scripts" {
										packageJson["scripts"].(map[string]interface{})[k3] = v3
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return defaults, defaultEnvs, projects, scripts, vars, packageJson, pipes
}

func ApplyPackageJSONOverrides(packageJSON *PackageJSON, packageJSONOverrides map[string]interface{}) *PackageJSON {
	for k, v := range packageJSONOverrides {
		vType := reflect.TypeOf(v)
		if vType.Kind() == reflect.Map {
			if k == "scripts" {
				if packageJSON.Scripts == nil {
					packageJSON.Scripts = make(map[string]string, 1000)
				}
				for k2, v2 := range v.(map[string]interface{}) {
					packageJSON.Scripts[k2] = v2.(string)
				}
			}
		}
	}
	return packageJSON
}

func ParseFlags() *FlagList {
	var flagList *FlagList
	flagList = new(FlagList)
	flagList.ExecuteAlias = flag.Bool("a", false, "Execute an alias")
	flagList.NoDefaultValues = flag.Bool("D", false, "Do not use default values")
	flagList.ShowScript = flag.Bool("s", false, "Show the script")
	flagList.ShowHelp = flag.Bool("h", false, "Show help")
	flagList.ShowList = flag.Bool("l", false, "Show all scripts")
	flagList.ShowLicense = flag.Bool("L", false, "Show licenses of dependencies")
	flagList.ShowVersion = flag.Bool("v", false, "Show current version")
	flagList.DummyCode = flag.Bool("d", false, "Exec some development dummy code")
	flagList.UseAnotherPath = flag.String("p", "", "Use another path to find the package.json")
	flagList.ShowCurrentProjectInfo = flag.Bool("i", false, "Show current project info")
	flagList.AddProject = flag.Bool("pa", false, "Add a project to the config")
	flagList.RemoveProject = flag.Bool("pr", false, "Remove a project from the config")
	flagList.GetProjectPath = flag.Bool("path", false, "Get the path of a project from the config")
	flagList.ListProjects = flag.Bool("pl", false, "List all projects from the config")
	flagList.BeVerbose = flag.Bool("V", false, "Be verbose, shows all environment variables set by nrun")
	flagList.PassthruNpm = flag.Bool("n", false, "Pass through to npm")
	flagList.SystemInfo = flag.Bool("I", false, "Get the system information")
	flagList.ExecuteCommand = flag.Bool("e", false, "Execute a command")
	flagList.ExecuteCommandInProjects = flag.Bool("ep", false, "Execute a command in all projects")
	flagList.ExecuteScript = flag.Bool("x", false, "Execute a script")
	flagList.ExecuteMultipleScripts = flag.Bool("xm", false, "Execute multiple scripts")
	flagList.ExecuteScriptInProjects = flag.Bool("xp", false, "Execute a script in all projects")
	flagList.ListExecutableScripts = flag.Bool("xl", false, "List all executable scripts")
	flagList.ShowExecutableScript = flag.String("xs", "", "Show an executable script")
	flagList.AddToExecutableScript = flag.String("xa", "", "Add to an executable script")
	flagList.RemoveExecutableScript = flag.String("xr", "", "Remove an executable script")
	flagList.MeasureTime = flag.Bool("T", false, "Measure the time it takes to execute the script")
	flagList.WebGet = flag.Bool("w", false, "Do a http(s) request from the web")
	flagList.WebGetTemplate = flag.String("wt", "", "Do a web request based on a template defined in the global .nrun.json")
	flagList.WebGetHeader = flag.Bool("wh", false, "Show headers for the web response")
	flagList.WebGetHeaderOnly = flag.Bool("who", false, "Show only headers for the web response")
	flagList.WebGetNoBody = flag.Bool("wnb", false, "Do not show the body for the web response")
	flagList.WebGetInformation = flag.Bool("wi", false, "Show information about the web response")
	flagList.WebGetMethod = flag.String("wm", "", "Set the method to use for the web request")
	flagList.WebGetFormat = flag.String("wf", "", "Set the format for the web request")
	flagList.WebGetAll = flag.Bool("wa", false, "Get information, headers and body for the web response")
	flagList.XAuthToken = flag.String("xat", "", "Set the X-AUTH-TOKEN to use")
	flagList.UnpackJWTToken = flag.Bool("jwt", false, "Unpack a JWT token")
	flagList.SignJWTToken = flag.Bool("jwt-sign", false, "Sign a JWT token")
	flagList.ValidateJWTToken = flag.String("jwt-validate", "", "Validate a JWT token")
	flagList.TellAJoke = flag.Bool("joke", false, "Tell a joke")
	flagList.VersionInformatrion = flag.Bool("vi", false, "Show version information")
	flagList.NoOverride = flag.Bool("no", false, "Do not override the default values")
	flagList.NoPackageJSONOverride = flag.Bool("npo", false, "Do not override the package.json")
	flagList.NoDefaultValues2 = flag.Bool("ndv", false, "Do not use default values")
	flagList.NoPipes = flag.Bool("np", false, "Do not use pipes")
	flagList.ForcePipes = flag.Bool("fp", false, "Force pipes")
	flagList.Sleep = flag.Int64("sleep", 0, "Sleep for a given amount of milliseconds")
	// Inactive flags
	flagList.TestAlarm = flag.Int64("t", 0, "Measure times in tests and notify when they are too long (time given in milliseconds)")

	usr, _ := user.Current()
	homeDir := usr.HomeDir
	config, _ := ReadConfig(homeDir + "/.nrun.json")
	if config != nil {
		if config.PersonalFlags != nil {
			flagList.PersonalFlags = make(map[string]*bool, len(config.PersonalFlags))
			for k, _ := range config.PersonalFlags {
				flagList.PersonalFlags[k] = flag.Bool(k, false, "Personal flag: "+k+". Usage is defined by the user in ~/.nrun.json")
			}
		}
	}
	flag.Parse()

	/* Override WebGetHeader and WebGetNoBody if WebGetHeaderOnly is set */
	if *flagList.WebGetHeaderOnly {
		*flagList.WebGetHeader = true
		*flagList.WebGetNoBody = true
	}

	return flagList
}

var notifyQueue = make(chan string, 25)
var NotifyWaitGroup sync.WaitGroup
var WaitingNotifications int32 = 0

func NotificationRunner() {
	sayPath, err := exec.LookPath("say")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var isActive bool = false
	// Infinite loop to wait for messages
	for {
		for {
			if !isActive {
				break
			}
			time.Sleep(time.Millisecond * 10)
		}
		isActive = true
		message := <-notifyQueue
		cmd := exec.Command(sayPath, []string{"--voice=Daniel", message}...)
		cmd.Start()
		go func() {
			atomic.AddInt32(&WaitingNotifications, -1)
			cmd.Wait()
			isActive = false
			NotifyWaitGroup.Done()
		}()
	}
}

func Notify(message string) {
	// Increment wait group
	NotifyWaitGroup.Add(1)
	// Increment number of waiting notifications
	atomic.AddInt32(&WaitingNotifications, 1)
	// Add message to queue
	notifyQueue <- message
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.Mode().IsDir()
}

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	if stat.Mode().IsDir() {
		return false
	}
	return stat.Mode().IsRegular()
}

func CopyFile(src string, dst string) error {
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

func Contains(stringsToSearch []string, key string) bool {
	for _, stringInList := range stringsToSearch {
		if strings.ToLower(stringInList) == key {
			return true
		}
	}
	return false
}

func WildMatch(stringsToSearch []string, key string) bool {
	for _, stringInList := range stringsToSearch {
		match, err := regexp.MatchString(".*"+strings.ToLower(stringInList)+".*", strings.ToLower(key))
		if match && err == nil {
			return true
		}
	}
	return false
}

func parseLine(raw string) (key string, value int) {
	text := strings.ReplaceAll(raw[:len(raw)-2], " ", "")
	keyValue := strings.Split(text, ":")
	intValue, _ := strconv.Atoi(keyValue[1])
	return keyValue[0], intValue * 1
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

var configCache = make(map[string]*Config, 100)

func WriteConfig(filename string, config *Config) error {
	configCache[filename] = config
	jsonFile, err := os.Create(filename)
	if err != nil {
		return err
	} else {
		jsonData, _ := json.MarshalIndent(config, "", "  ")
		_, err = jsonFile.Write(jsonData)
		if err != nil {
			return err
		}
		err := jsonFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadConfig(filepath string) (*Config, error) {
	if configCache[filepath] != nil {
		return configCache[filepath], nil
	}
	if !FileExists(filepath) {
		return nil, errors.New("config file not found")
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if _, err := os.Stat(filepath); !errors.Is(err, os.ErrNotExist) {
		jsonFile, err := os.Open(filepath)
		if err != nil {
			log.Println("Failed with", err)
			return nil, err
		} else {
			byteValue, _ := os.ReadFile(jsonFile.Name())
			var config Config
			_ = json.Unmarshal(byteValue, &config)
			jsonFile.Close()
			configCache[filepath] = &config
			return &config, nil
		}
	}
	return nil, err
}

func ProcessVarsOnString(data []byte, vars map[string]string) []byte {
	for key, value := range vars {
		data = []byte(strings.Replace(string(data), "{{"+key+"}}", value, -1))
	}
	return data
}
func ApplyVars(data map[string]string, vars map[string]string) map[string]string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return data
	}

	jsonData = ProcessVarsOnString(jsonData, vars)

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return data
	}
	return data
}
func ApplyVarsArray(data map[string][]string, vars map[string]string) map[string][]string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return data
	}

	jsonData = ProcessVarsOnString(jsonData, vars)

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return data
	}
	return data
}
func ApplyVarsTemplateStruct(data WebGetTemplateStruct, vars map[string]string) WebGetTemplateStruct {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		return data
	}

	jsonData = ProcessVarsOnString(jsonData, vars)

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Println("Error:", err)
		return data
	}
	return data
}
