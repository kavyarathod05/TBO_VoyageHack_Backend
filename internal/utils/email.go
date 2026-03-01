package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
)

// ResendPayload defines the body structure for Resend API
type ResendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendEmail sends an email using Resend HTTPS API (bypasses SMTP restrictions)
func SendEmail(cfg *config.Config, to []string, subject string, body string) error {
	apiKey := cfg.ResendAPIKey
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY missing - cannot send email")
	}

	// Use Resend's onboarding email if no custom domain is verified
	// Replaced "tboemailservice@gmail.com" with a valid Resend sender
	from := "TBO <onboarding@resend.dev>"

	payload := ResendPayload{
		From:    from,
		To:      to,
		Subject: subject,
		HTML:    body,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Resend API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend API error (Status %d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("📧 Email sent to %v via Resend API", to)
	return nil
}
