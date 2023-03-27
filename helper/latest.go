package helper

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type NodeJSRelease struct {
	Version  string      `json:"version"`
	Date     string      `json:"date"`
	Files    []string    `json:"files"`
	Npm      string      `json:"npm,omitempty"`
	V8       string      `json:"v8"`
	Uv       string      `json:"uv,omitempty"`
	Zlib     string      `json:"zlib,omitempty"`
	Openssl  string      `json:"openssl,omitempty"`
	Modules  string      `json:"modules,omitempty"`
	Lts      interface{} `json:"lts"`
	Security bool        `json:"security"`
}

type NodeJSReleases []NodeJSRelease

func VersionCompare(v1, v2 string) int {
	// Compare two version strings
	// Returns 1 if v1 > v2
	// Returns 0 if v1 == v2
	// Returns -1 if v1 < v2
	// Returns -2 if v1 or v2 is not a valid version string
	// v1 and v2 must be in the format "vX.Y.Z" or "X.Y.Z"
	// where X, Y and Z are integers
	// e.g. "v1.2.3" or "1.2.3"
	// The "v" prefix is optional
	// The version strings must have the same number of components
	// e.g. "v1.2.3" and "v1.2.3.4" are not valid
	// e.g. "v1.2.3" and "v1.2" are not valid
	// e.g. "v1.2.3" and "v1.2.3.0" are valid
	// e.g. "v1.2.3" and "v1.2.3" are valid
	// e.g. "v1.2.3" and "v1.
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")
	for i := 0; i < len(v1Parts); i++ {
		if i >= len(v2Parts) {
			return -2
		}
		v1Part := v1Parts[i]
		v2Part := v2Parts[i]
		if v1Part[0] == 'v' {
			v1Part = v1Part[1:]
		}
		if v2Part[0] == 'v' {
			v2Part = v2Part[1:]
		}
		v1Int, err := strconv.Atoi(v1Part)
		if err != nil {
			return -2
		}
		v2Int, err := strconv.Atoi(v2Part)
		if err != nil {
			return -2
		}
		if v1Int > v2Int {
			return 1
		}
		if v1Int < v2Int {
			return -1
		}
	}
	return 0
}

func GetLatestNodeJSRelease() (NodeJSReleases, error) {
	var releases NodeJSReleases
	// Read data from the URL
	resp, err := http.Get("https://nodejs.org/download/release/index.json")
	if err != nil {
		return releases, err
	}
	defer resp.Body.Close()
	// Decode the JSON data
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return releases, err
	}
	return releases, err
}

func GetLatestNVMRelease() (string, error) {
	var latest string
	// Read data from the URL
	resp, err := http.Get("https://api.github.com/repos/nvm-sh/nvm/releases/latest")
	if err != nil {
		return latest, err
	}
	defer resp.Body.Close()
	// Decode the JSON data
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return latest, err
	}
	latest = data["tag_name"].(string)
	return latest, nil
}
