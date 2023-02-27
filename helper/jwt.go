package helper

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
)

func UnpackJWTToken(token string) error {
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return errors.New("Invalid JWT token")
	}
	// Decode the first part of the token
	decodedFirst, err1 := base64.RawURLEncoding.DecodeString(tokenParts[0])
	if err1 != nil {
		return err1
	}
	// Decode the second part of the token
	decodedSecond, err2 := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err2 != nil {
		return err2
	}
	// Print the decoded parts
	// Prettify the JSON
	var data interface{}
	json.Unmarshal(decodedFirst, &data)
	pretty1, err3 := json.MarshalIndent(data, "", "    ")
	if err3 != nil {
		return err3
	}
	println("Header:", string(pretty1))
	json.Unmarshal(decodedSecond, &data)
	pretty2, err4 := json.MarshalIndent(data, "", "    ")
	if err4 != nil {
		return err4
	}
	println("Payload:", string(pretty2))
	return nil
}

func SignJWTToken(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("Invalid number of arguments")
	}
	secret := args[0]
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		header := base64.URLEncoding.EncodeToString([]byte("{\"alg\":\"HS256\",\"typ\":\"JWT\"}"))
		header = strings.ReplaceAll(header, "=", "")
		payload := base64.URLEncoding.EncodeToString(stdin)
		payload = strings.ReplaceAll(payload, "=", "")
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(header + "." + payload))
		signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
		signature = strings.ReplaceAll(signature, "=", "")
		fmt.Println(header + "." + payload + "." + signature)
	} else if len(args) == 2 {
		usr, _ := user.Current()
		dir := usr.HomeDir
		filename := args[1]

		if FileExists(filename) == false {
			config, _ := ReadConfig(dir + "/.nrun.json")
			if config.TokenTemplates[filename] == "" {
				return errors.New("Token template not found")
			}
		}
		header := base64.URLEncoding.EncodeToString([]byte("{\"alg\":\"HS256\",\"typ\":\"JWT\"}"))
		header = strings.ReplaceAll(header, "=", "")
		file, err := os.Open(filename)
		if err != nil {
			return errors.New("Token template file not found")
		}
		defer file.Close()
		var payload []byte
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			payload = append(payload, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		payload = []byte(base64.URLEncoding.EncodeToString(payload))
		payload = []byte(strings.ReplaceAll(string(payload), "=", ""))
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(header + "." + string(payload)))
		signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
		signature = strings.ReplaceAll(signature, "=", "")
		fmt.Println(header + "." + string(payload) + "." + signature)
	} else {
		fmt.Println("No data provided")
		return errors.New("No data provided")
	}
	return nil
	return errors.New("Not implemented yet")
}

func ValidateJWTToken(flagList *FlagList, args []string) error {
	if len(args) != 1 {
		return errors.New("Invalid number of arguments")
	}
	secret := args[0]
	token := *flagList.ValidateJWTToken
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return errors.New("Invalid JWT token")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(tokenParts[0] + "." + tokenParts[1]))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	signature = strings.ReplaceAll(signature, "=", "")
	if signature == tokenParts[2] {
		fmt.Println("Valid JWT token")
	} else {
		fmt.Println("Invalid JWT token")
		fmt.Println("Using secret: " + secret)
	}
	return nil
}
