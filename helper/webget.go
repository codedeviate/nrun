package helper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os/user"
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

	ProcessWebRequest(flagList, args[0], method, body, headers, requestStart)
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
	ProcessWebRequest(flagList, template.URL, template.Method, template.Body, template.Headers, requestStart)
}

func ProcessWebRequest(flagList *FlagList, url string, method string, body string, reqHeaders map[string]string, requestStart time.Time) {
	data, headers, response, err := DoWebRequest(url, method, body, reqHeaders)
	if err != nil {
		fmt.Println("Failed to do web request:", err)
		return
	}

	if flagList.WebGetInformation != nil && *flagList.WebGetInformation {
		fmt.Println("URL:", url)
		fmt.Println("Method:", method)
		fmt.Println("Status:", response.Status)
		fmt.Println("Size:", len(data), "bytes")
		fmt.Println("Time:", time.Since(requestStart).Milliseconds(), "ms")
		fmt.Println("X-Auth-Token:", reqHeaders["X-Auth-Token"])
		fmt.Println("")
	}

	if (flagList.WebGetHeader != nil && *flagList.WebGetHeader) || (flagList.WebGetInformation != nil && *flagList.WebGetInformation) {
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

	if (flagList.WebGetNoBody != nil && *flagList.WebGetNoBody) || (flagList.WebGetInformation != nil && *flagList.WebGetInformation) {
		return
	}

	format := *flagList.WebGetFormat
	if len(format) == 0 {
		format = "json"
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
