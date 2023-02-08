package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func ShowLicense(path string, script string, args []string) {
	licenseList := make(map[string][]string, 1000)
	licenseList = ShowLicenseInfo(path, licenseList)
	licenseListKeys := make([]string, 0, len(licenseList))
	for k := range licenseList {
		licenseListKeys = append(licenseListKeys, k)
	}
	sort.Strings(licenseListKeys)
	if len(script) > 0 {
		args = append(args, script)
	}
	for index, key := range licenseListKeys {
		values := licenseList[licenseListKeys[index]]
		if len(args) == 0 || script == "names" || (Contains(args, strings.ToLower(key)) || Contains(args, "names") || WildMatch(args, key)) {
			if len(args) == 0 || script == "names" || (len(args) == 1 && Contains(args, "names")) || (len(args) > 1 && Contains(args, "names") && WildMatch(args, key)) || !Contains(args, "names") {
				fmt.Println(key)
			}
			if !Contains(args, "names") && script != "names" {
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
}

func ShowLicenseInfo(path string, licenseList LicenseList) LicenseList {
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
								licenseList = ShowLicenseInfo(path+"/node_modules/"+file.Name(), licenseList)
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
						licenseList = ShowLicenseInfo(path+"/"+file.Name(), licenseList)
					}
				}
			}
		}
	}
	return licenseList
}
