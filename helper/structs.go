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
	Env                 map[string]map[string]string    `json:"env"`
	Path                map[string]map[string]string    `json:"path"`
	Vars                map[string]string               `json:"vars"`
	Projects            map[string]string               `json:"projects"`
	Alias               map[string]string               `json:"alias"`
	Scripts             map[string][]string             `json:"scripts"`
	WebGetTemplates     map[string]WebGetTemplateStruct `json:"webget"`
	XAuthTokens         map[string]string               `json:"xauthtokens"`
	PersonalFlags       map[string][]string             `json:"personalflags"`
	TokenTemplates      map[string]string               `json:"tokentemplates"`
	PackageJSONOverride map[string]interface{}          `json:"package.json"`
}

type WebGetTemplateStruct struct {
	Method     string                 `json:"method"`
	URL        string                 `json:"url"`
	Format     string                 `json:"format"`
	Body       string                 `json:"body"`
	Headers    map[string]string      `json:"headers"`
	XAuthToken string                 `json:"xauthtoken"`
	Flags      map[string]interface{} `json:"flags"`
}

type LicenseList map[string][]string
type FlagList struct {
	ExecuteAlias             *bool
	NoDefaultValues          *bool
	ShowScript               *bool
	ShowHelp                 *bool
	ShowList                 *bool
	ShowLicense              *bool
	ShowVersion              *bool
	DummyCode                *bool
	UseAnotherPath           *string
	UsedPath                 string
	OriginalPath             string
	ShowCurrentProjectInfo   *bool
	AddProject               *bool
	RemoveProject            *bool
	GetProjectPath           *bool
	ListProjects             *bool
	BeVerbose                *bool
	PassthruNpm              *bool
	SystemInfo               *bool
	ExecuteCommand           *bool
	ExecuteCommandInProjects *bool
	ExecuteScript            *bool
	ExecuteMultipleScripts   *bool
	ExecuteScriptInProjects  *bool
	ListExecutableScripts    *bool
	ShowExecutableScript     *string
	AddToExecutableScript    *string
	RemoveExecutableScript   *string
	MeasureTime              *bool
	WebGet                   *bool
	WebGetTemplate           *string
	WebGetHeader             *bool
	WebGetHeaderOnly         *bool
	WebGetNoBody             *bool
	WebGetInformation        *bool
	WebGetAll                *bool
	WebGetMethod             *string
	WebGetFormat             *string
	XAuthToken               *string
	TestAlarm                *int64 // Time in milliseconds. Currently not used
	Vars                     map[string]string
	PersonalFlags            map[string]*bool
	UnpackJWTToken           *bool
	SignJWTToken             *bool
	ValidateJWTToken         *string
}

type Memory struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
}

type JWTToken struct {
	Header    JWTTokenHeader
	Payload   JWTTokenPayload
	Signature string
}
type JWTTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}
type JWTTokenPayload struct {
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Iss string `json:"iss"`
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Nbf int64  `json:"nbf"`
	Jti string `json:"jti"`
}
