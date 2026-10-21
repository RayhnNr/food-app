package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"
)

type recaptchaResponse struct {
	Success bool `json:"success"`
}

func verifyCaptcha(token, ip string) bool {
	if token == "" {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {os.Getenv("RECAPTCHA_SECRET")},
			"response": {token},
			"remoteip": {ip},
		},
	)

	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result recaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}
	return  result.Success
}