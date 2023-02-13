package helper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

func WebGet(args []string, flagList *FlagList) {
	requestStart := time.Now()

	method := "GET"
	if flagList.WebGetMethod != nil && len(*flagList.WebGetMethod) > 0 {
		method = *flagList.WebGetMethod
	}
	body := ""
	headers := make(map[string]string)
	if flagList.XAuthToken != nil && len(*flagList.XAuthToken) > 0 {
		usr, _ := user.Current()
		dir := usr.HomeDir
		config, err := ReadConfig(dir + "/.nrun.json")

		if err != nil {
			fmt.Println("Failed to read config:", err)
			return
		}

		if _, ok := config.XAuthTokens[*flagList.XAuthToken]; !ok {
			headers["X-Auth-Token"] = *flagList.XAuthToken
			return
		} else {
			headers["X-Auth-Token"] = config.XAuthTokens[*flagList.XAuthToken]
		}
	}

	if flagList.WebGetFormat == nil || len(*flagList.WebGetFormat) == 0 {
		*flagList.WebGetFormat = "auto"
	}

	// Check if we have data in the stdin pipe to use as body
	fi, statErr := os.Stdin.Stat()
	if statErr == nil && (fi.Mode()&os.ModeCharDevice == 0) {
		stdin, inErr := io.ReadAll(os.Stdin)
		if inErr == nil && len(stdin) > 0 {
			body = string(stdin)
		}
	}

	if len(args) > 0 {
		ProcessWebRequest(flagList, args[0], method, body, headers, requestStart)
	} else {
		log.Println("No URL given", args)
	}
}

func WebGetTemplate(args []string, flagList *FlagList) {
	usr, _ := user.Current()
	dir := usr.HomeDir

	requestStart := time.Now()

	config, err := ReadConfig(dir + "/.nrun.json")

	if err != nil {
		fmt.Println("Failed to read config:", err)
		return
	}

	var template WebGetTemplateStruct
	var ok bool

	if method := *flagList.WebGetMethod; len(method) > 0 {
		template.Method = method
	}

	if template, ok = config.WebGetTemplates[*flagList.WebGetTemplate]; !ok {
		fmt.Println("Template", *flagList.WebGetTemplate, "does not exist")
		return
	}

	template = ApplyVarsTemplateStruct(template, flagList.Vars)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if template.Headers == nil {
		template.Headers = make(map[string]string)
	}

	if len(template.XAuthToken) > 0 {
		template.Headers["X-Auth-Token"] = template.XAuthToken
	}

	if flagList.XAuthToken != nil && len(*flagList.XAuthToken) > 0 {
		if _, ok := config.XAuthTokens[*flagList.XAuthToken]; !ok {
			fmt.Println("X-Auth-Token", "\""+*flagList.XAuthToken+"\"", "does not exist")
			return
		} else {
			template.Headers["X-Auth-Token"] = config.XAuthTokens[*flagList.XAuthToken]
		}
	}

	if len(args) > 0 {
		template.URL = args[0]
	}

	if flagList.WebGetMethod != nil && len(*flagList.WebGetMethod) > 0 {
		template.Method = *flagList.WebGetMethod
	}

	if flagList.WebGetFormat == nil || len(*flagList.WebGetFormat) == 0 {
		*flagList.WebGetFormat = template.Format
	}

	if template.Flags != nil {
		if template.Flags["wi"] != nil {
			*flagList.WebGetInformation = template.Flags["wi"].(bool)
		}
		if template.Flags["wa"] != nil {
			*flagList.WebGetAll = template.Flags["wa"].(bool)
		}
		if template.Flags["wh"] != nil {
			*flagList.WebGetHeader = template.Flags["wh"].(bool)
		}
		if template.Flags["who"] != nil {
			*flagList.WebGetHeaderOnly = template.Flags["who"].(bool)
		}
		if template.Flags["wnb"] != nil {
			*flagList.WebGetNoBody = template.Flags["wnb"].(bool)
		}
		if template.Flags["wm"] != nil {
			*flagList.WebGetMethod = template.Flags["wm"].(string)
		}
		if template.Flags["wf"] != nil {
			*flagList.WebGetFormat = template.Flags["wf"].(string)
		}
	}

	if template.Body != "" {
		if template.Body[0:2] == "@@" {
			template.Body = template.Body[2:]
			bodyParts := strings.Split(template.Body, ":")
			if strings.TrimSpace(bodyParts[0]) == "file" {
				if len(bodyParts) > 1 {
					fileName := strings.TrimSpace(bodyParts[1])
					if fileName[0] == '~' {
						fileName = dir + "/" + fileName
					} else if fileName[0] == '.' {
						cwd, _ := os.Getwd()
						fileName = cwd + "/" + fileName
					} else if fileName[0] != '/' {
						fileName = flagList.UsedPath + "/" + fileName
					}
					fileName = strings.ReplaceAll(fileName, "//", "/")
					fileName, err = filepath.Abs(fileName)
					if err != nil {
						log.Println("Failed to get absolute path for file:", err)
						return
					}
					file, err := os.Open(fileName)
					if err != nil {
						log.Println("Failed to open file:", err)
						return
					}
					defer file.Close()
					body, err := io.ReadAll(file)
					if err != nil {
						log.Println("Failed to read file:", err)
						return
					}
					template.Body = string(body)
				} else {
					log.Println("No file name specified for file body")
				}
			}
		}
	}

	ProcessWebRequest(flagList, template.URL, template.Method, template.Body, template.Headers, requestStart)
}

func ProcessWebRequest(flagList *FlagList, url string, method string, body string, reqHeaders map[string]string, requestStart time.Time) {
	data, headers, response, err := DoWebRequest(url, method, body, reqHeaders)
	if err != nil {
		fmt.Println("Failed to do web request:", err)
		return
	}

	showInformation := flagList.WebGetInformation != nil && *flagList.WebGetInformation
	showInformation = showInformation || (flagList.WebGetAll != nil && *flagList.WebGetAll)

	showHeader := flagList.WebGetHeader != nil && *flagList.WebGetHeader
	showHeader = showHeader || (flagList.WebGetHeaderOnly != nil && *flagList.WebGetHeaderOnly)
	showHeader = showHeader || (flagList.WebGetInformation != nil && *flagList.WebGetInformation)
	showHeader = showHeader || (flagList.WebGetAll != nil && *flagList.WebGetAll)

	showBody := !(flagList.WebGetHeaderOnly != nil && *flagList.WebGetHeaderOnly)
	showBody = showBody && !(flagList.WebGetInformation != nil && *flagList.WebGetInformation)
	showBody = showBody && !(flagList.WebGetNoBody != nil && *flagList.WebGetNoBody)
	showBody = showBody || (flagList.WebGetAll != nil && *flagList.WebGetAll)

	if showInformation {
		fmt.Println("URL:", url)
		fmt.Println("Method:", method)
		fmt.Println("Status:", response.Status)
		fmt.Println("Size:", len(data), "bytes")
		fmt.Println("Time:", time.Since(requestStart).Milliseconds(), "ms")
		fmt.Println("X-Auth-Token:", reqHeaders["X-Auth-Token"])
		fmt.Println("")
	}

	if showHeader {
		fmt.Println("Headers:")
		for key, value := range headers {
			if reflect.TypeOf(value) == reflect.TypeOf([]string{}) {
				fmt.Println("  "+key+":", strings.Join(value, ", "))
				continue
			}
			fmt.Println("  "+key+":", value)
		}
		if len(response.Cookies()) > 0 {
			fmt.Println("Cookies:")
			for _, cookie := range response.Cookies() {
				fmt.Println("  "+cookie.Name+":", cookie.Value)
			}
		}
		fmt.Println("")
	}

	if !showBody {
		return
	}

	format := *flagList.WebGetFormat
	if len(format) == 0 {
		format = "auto"
	}
	format = strings.ToLower(format)

	cZero := string(data[0])
	if cZero == "{" || cZero == "[" {
		if format == "auto" {
			format = "json"
		}
	} else if cZero == "<" {
		cOne := string(data[1])
		if cOne == "?" {
			if format == "auto" {
				format = "xml"
			}
		} else if cOne == "!" {
			if format == "auto" {
				format = "html"
			}
		} else {
			if format == "auto" {
				format = "text"
			}
		}
	} else {
		if format == "auto" {
			format = "text"
		}
	}

	if format == "json" {
		var jsonData interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			fmt.Println("Failed to parse JSON:", err)
			return
		}
		prettyJson, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			fmt.Println("Failed to parse JSON:", err)
			return
		}
		fmt.Println(string(prettyJson))
	} else if format == "text" {
		fmt.Println(string(data))
	} else if format == "raw" {
		fmt.Println(data)
	} else if format == "base64" {
		text := base64.StdEncoding.EncodeToString(data)
		fmt.Println(text)
	} else if format == "hex" {
		text := hex.EncodeToString(data)
		fmt.Println(text)
	} else if format == "xml" {
		var xmlData interface{}
		err = xml.Unmarshal(data, &xmlData)
		if err != nil {
			fmt.Println("Failed to parse XML:", err)
			return
		}
		prettyXml, err := xml.MarshalIndent(xmlData, "", "  ")
		if err != nil {
			fmt.Println("Failed to parse XML:", err)
			return
		}
		fmt.Println(string(prettyXml))
	} else {
		fmt.Println("Format", format, "not supported")
		fmt.Println(data)
	}
}

func DoWebRequest(url string, method string, body string, headers map[string]string) ([]byte, http.Header, http.Response, error) {
	client := http.Client{}
	bodyReader := strings.NewReader(body)
	method = strings.ToUpper(method)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return nil, nil, http.Response{}, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return nil, nil, http.Response{}, err
	}
	defer response.Body.Close()
	bodyResponse, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read response:", err)
		return nil, nil, http.Response{}, err
	}
	return bodyResponse, response.Header.Clone(), *response, nil
}
