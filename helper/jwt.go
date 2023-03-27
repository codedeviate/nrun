package helper

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"time"
)

func UnpackJWTToken(token string) error {
	// Check if the token is empty
	if len(token) == 0 {
		// Check if the token is being piped in
		fi, statErr := os.Stdin.Stat()
		if statErr == nil && (fi.Mode()&os.ModeCharDevice == 0) {
			stdin, inErr := io.ReadAll(os.Stdin)
			if inErr == nil && len(stdin) > 0 {
				token = string(stdin)
			}
		}
	}
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
	payload := JWTTokenPayload{}
	json.Unmarshal(decodedSecond, &payload)
	extraInfo := ""
	if payload.Iat != 0 {
		extraInfo += fmt.Sprintf("Issued at: %s\n", time.Unix(payload.Iat, 0).Format(time.RFC3339))
	}
	if payload.Nbf != 0 {
		extraInfo += fmt.Sprintf("Not before: %s", time.Unix(payload.Nbf, 0).Format(time.RFC3339))
		if payload.Nbf > time.Now().Unix() {
			extraInfo += " (not yet valid)\n"
		} else {
			extraInfo += "\n"
		}
	}
	if payload.Exp != 0 {
		extraInfo += fmt.Sprintf("Expires at: %s", time.Unix(payload.Exp, 0).Format(time.RFC3339))
		if payload.Exp < time.Now().Unix() {
			extraInfo += " (expired)\n"
		} else {
			extraInfo += "\n"
		}
	}
	if payload.Iss != "" {
		extraInfo += fmt.Sprintf("Issuer: %s\n", payload.Iss)
	}
	if payload.Aud != "" {
		extraInfo += fmt.Sprintf("Audience: %s\n", payload.Aud)
	}
	if payload.Sub != "" {
		extraInfo += fmt.Sprintf("Subject: %s\n", payload.Sub)
	}
	if payload.Jti != "" {
		extraInfo += fmt.Sprintf("JWT ID: %s\n", payload.Jti)
	}
	if extraInfo != "" {
		println("\nExtra info:\n" + extraInfo)
	}
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
		// Replace dynamic parts of the token
		stdin = []byte(strings.ReplaceAll(string(stdin), "{{NOW}}", fmt.Sprintf("%d", time.Now().Unix())))
		stdin = []byte(strings.ReplaceAll(string(stdin), "{{NOW+1H}}", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix())))
		stdin = []byte(strings.ReplaceAll(string(stdin), "{{NOW+1D}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24).Unix())))
		stdin = []byte(strings.ReplaceAll(string(stdin), "{{NOW+1W}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24*7).Unix())))
		stdin = []byte(strings.ReplaceAll(string(stdin), "{{NOW+1M}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24*30).Unix())))
		stdin = []byte(strings.ReplaceAll(string(stdin), "{{NOW+1Y}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24*365).Unix())))
		// Build the header
		header := base64.URLEncoding.EncodeToString([]byte("{\"alg\":\"HS256\",\"typ\":\"JWT\"}"))
		header = strings.ReplaceAll(header, "=", "")
		// Build the payload
		payload := base64.URLEncoding.EncodeToString(stdin)
		payload = strings.ReplaceAll(payload, "=", "")
		// Sign the token
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(header + "." + payload))
		signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
		signature = strings.ReplaceAll(signature, "=", "")
		// Print the token
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
		// Build the header
		header := base64.URLEncoding.EncodeToString([]byte("{\"alg\":\"HS256\",\"typ\":\"JWT\"}"))
		header = strings.ReplaceAll(header, "=", "")
		// Read the payload
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
		// Replace dynamic parts of the token
		payload = []byte(strings.ReplaceAll(string(payload), "{{NOW}}", fmt.Sprintf("%d", time.Now().Unix())))
		payload = []byte(strings.ReplaceAll(string(payload), "{{NOW+1H}}", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix())))
		payload = []byte(strings.ReplaceAll(string(payload), "{{NOW+1D}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24).Unix())))
		payload = []byte(strings.ReplaceAll(string(payload), "{{NOW+1W}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24*7).Unix())))
		payload = []byte(strings.ReplaceAll(string(payload), "{{NOW+1M}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24*30).Unix())))
		payload = []byte(strings.ReplaceAll(string(payload), "{{NOW+1Y}}", fmt.Sprintf("%d", time.Now().Add(time.Hour*24*365).Unix())))
		// Build the payload
		payload = []byte(base64.URLEncoding.EncodeToString(payload))
		payload = []byte(strings.ReplaceAll(string(payload), "=", ""))
		// Sign the token
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

	header := JWTTokenHeader{}
	decodedHeader, errHeader := base64.RawURLEncoding.DecodeString(tokenParts[0])
	if errHeader != nil {
		return errHeader
	}
	json.Unmarshal(decodedHeader, &header)
	if header.Alg != "HS256" {
		fmt.Println("JWT token algorithm not supported")
		return nil
	}
	if header.Typ != "JWT" {
		fmt.Println("JWT token type not supported")
		return nil
	}

	payload := JWTTokenPayload{}
	decodedPayload, errPayload := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if errPayload != nil {
		return errPayload
	}
	json.Unmarshal(decodedPayload, &payload)
	if payload.Exp > 0 && payload.Exp < time.Now().Unix() {
		fmt.Println("JWT token expired")
		return nil
	}
	if payload.Nbf > time.Now().Unix() {
		fmt.Println("JWT token not valid yet")
		return nil
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
