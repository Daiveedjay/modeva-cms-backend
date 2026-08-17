package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// CustomerEmailData holds data for customer email
type CustomerEmailData struct {
	CustomerName  string
	CustomerEmail string
	Subject       string
	Message       string
}

// SendCustomerEmail sends a custom email to a customer via Resend
func (r *ResendClient) SendCustomerEmail(data CustomerEmailData) error {
	// HTML email template with inline styles
	htmlBody := r.buildCustomerEmailHTML(data)

	// Prepare request payload
	payload := map[string]interface{}{
		"from":    r.from,
		"to":      data.CustomerEmail,
		"subject": data.Subject,
		"html":    htmlBody,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[resend] failed to marshal payload: %v", err)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Make request to Resend API
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("[resend] failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[resend] failed to send request: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[resend] failed to read response: %v", err)
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("[resend] api returned status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("resend api error: status %d", resp.StatusCode)
	}

	log.Printf("[resend] customer email sent to %s", data.CustomerEmail)
	return nil
}

// buildCustomerEmailHTML creates a beautiful HTML body for customer email with inline styles
func (r *ResendClient) buildCustomerEmailHTML(data CustomerEmailData) string {
	// Convert plain text message to HTML paragraphs
	messageParagraphs := ""
	lines := strings.Split(data.Message, "\n\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			messageParagraphs += fmt.Sprintf(`<p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px; margin-top: 0;">%s</p>`, strings.TrimSpace(line))
		}
	}

	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body style="margin: 0; padding: 0; box-sizing: border-box; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif; background-color: #ffffff; color: #1a1a1a; line-height: 1.6;">
    <div style="background-color: #ffffff; padding: 60px 20px;">
      <div style="max-width: 600px; margin: 0 auto; background: #ffffff;">
        <!-- Header -->
        <div style="padding: 0 0 80px 0; text-align: left; position: relative;">
          <div style="font-size: 24px; font-weight: 700; color: #1a1a1a; letter-spacing: -0.3px; margin-bottom: 0;">Modeva</div>
        </div>

        <!-- Content -->
        <div style="padding: 0;">
          <p style="font-size: 36px; font-weight: 700; color: #000000; margin-bottom: 24px; letter-spacing: -0.8px; line-height: 1.2; margin-top: 0;">%s</p>

          <p style="font-size: 17px; color: #626262; line-height: 1.8; margin-bottom: 40px; margin-top: 0;">
            Hi <span style="color: #000000; font-weight: 600;">%s</span>,
          </p>

          %s

          <hr style="border: 0; height: 1px; background: #e5e5e5; margin: 60px 0;" />

          <p style="font-size: 13px; color: #626262; line-height: 1.7; margin-top: 40px;">
            If you have any questions or need assistance, don't hesitate to reach out to our support team.
          </p>
        </div>

        <!-- Footer -->
        <div style="padding: 40px 0 0 0; text-align: left;">
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin-bottom: 8px; margin-top: 0;">© 2026 Modeva. All rights reserved.</p>
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin-top: 0;">
            Questions?
            <a href="mailto:support@modeva.biz" style="color: #0066cc; text-decoration: none; font-size: 13px; font-weight: 500;">Contact support</a>
          </p>
        </div>
      </div>
    </div>
  </body>
</html>`, data.Subject, data.Subject, data.CustomerName, messageParagraphs)
}
