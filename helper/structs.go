package helper

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
	ShowScript               *bool
	ShowHelp                 *bool
	ShowList                 *bool
	ShowLicense              *bool
	ShowVersion              *bool
	DummyCode                *bool
	UseAnotherPath           *string
	UsedPath                 string
	ShowCurrentProjectInfo   *bool
	AddProject               *bool
	RemoveProject            *bool
	ListProjects             *bool
	BeVerbose                *bool
	PassthruNpm              *bool
	SystemInfo               *bool
	ExecuteCommand           *bool
	ExecuteCommandInProjects *bool
	ExecuteScript            *bool
	ExecuteScriptInProjects  *bool
	ListExecutableScripts    *bool
	ShowExecutableScript     *string
	AddToExecutableScript    *string
	RemoveExecutableScript   *string
	MeasureTime              *bool
}

type Memory struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
}
