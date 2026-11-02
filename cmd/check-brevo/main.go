// check-brevo checks API authentication without sending an email or exposing keys.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func check(label, key string) {
	fmt.Printf("%s: key_present=%t, surrounding_whitespace=%t, smtp_key=%t\n",
		label, key != "", key != strings.TrimSpace(key), strings.HasPrefix(key, "xsmtpsib-"))
	if key == "" {
		return
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.brevo.com/v3/account", nil)
	if err != nil {
		fmt.Println("Could not create authentication check")
		return
	}
	req.Header.Set("api-key", key)
	req.Header.Set("accept", "application/json")
	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		// Do not print headers, credentials, or account details.
		fmt.Println("Network/authentication check could not complete:", strings.ReplaceAll(err.Error(), key, "[REDACTED]"))
		return
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	fmt.Printf("%s: Brevo GET /v3/account HTTP %d\n", label, response.StatusCode)
	if response.StatusCode >= 400 {
		var detail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &detail) == nil {
			message := strings.ReplaceAll(detail.Message, key, "[REDACTED]")
			if len(message) > 500 {
				message = message[:500]
			}
			fmt.Printf("Brevo error: %s\n", message)
		}
	}
}

func main() {
	inherited := os.Getenv("BREVO_API_KEY")
	values, fileErr := godotenv.Read(".env")
	// Match main.go: existing process variables take precedence over .env.
	loadErr := godotenv.Load()
	effective := os.Getenv("BREVO_API_KEY")
	fmt.Printf("dotenv_readable=%t, dotenv_loaded=%t, inherited_key_present=%t\n", fileErr == nil, loadErr == nil, inherited != "")
	check("backend-effective-key", effective)
	if fileErr == nil && values["BREVO_API_KEY"] != "" && values["BREVO_API_KEY"] != effective {
		fmt.Println("Process environment overrides a different key in .env")
		check("dotenv-key", values["BREVO_API_KEY"])
	}
	fmt.Println("No emails sent; no database requests made.")
}
