package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// ResendClient handles email sending via Resend API
type ResendClient struct {
	apiKey string
	from   string
}

// NewResendClient creates a new Resend client
func NewResendClient() *ResendClient {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Fatal("RESEND_API_KEY environment variable not set")
	}

	from := os.Getenv("RESEND_FROM_EMAIL")
	if from == "" {
		from = "noreply@contact.modeva.shop" // Default from address
	}

	return &ResendClient{
		apiKey: apiKey,
		from:   from,
	}
}

// AdminInviteEmailData holds data for admin invite email
type AdminInviteEmailData struct {
	AdminName  string
	AdminEmail string
	InviteLink string
}

// SendAdminInviteEmail sends an admin invite email via Resend
func (r *ResendClient) SendAdminInviteEmail(data AdminInviteEmailData) error {
	// HTML email template with inline styles
	htmlBody := r.buildAdminInviteHTML(data)

	// Prepare request payload
	payload := map[string]interface{}{
		"from":    r.from,
		"to":      data.AdminEmail,
		"subject": "You've been invited to join Modeva Admin",
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

	log.Printf("[resend] admin invite email sent to %s", data.AdminEmail)
	return nil
}

// buildAdminInviteHTML creates a beautiful HTML body for admin invite email with inline styles
func (r *ResendClient) buildAdminInviteHTML(data AdminInviteEmailData) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Join Modeva Admin Team</title>
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
          <p style="font-size: 36px; font-weight: 700; color: #000000; margin-bottom: 24px; letter-spacing: -0.8px; line-height: 1.2; margin-top: 0;">Join our admin team</p>

          <p style="font-size: 17px; color: #626262; line-height: 1.8; margin-bottom: 40px; margin-top: 0;">
            <span style="color: #000000; font-weight: 600;">%s</span>, you've been invited to become a Modeva admin. You'll have the power to create, update, and manage tasks that keep operations running smoothly.
          </p>

          <div style="margin: 40px 0;">
            <div style="font-size: 12px; font-weight: 600; color: #626262; text-transform: uppercase; letter-spacing: 0.8px; margin-bottom: 16px;">What You Can Do</div>
            <div style="font-size: 17px; color: #1a1a1a; line-height: 1.8;">
              Create new tasks, update existing ones, and delete completed items. Everything you need to keep your workflow organized and efficient.
            </div>
          </div>

          <div style="text-align: left; margin: 50px 0 60px 0;">
            <a href="%s" style="display: inline-block; padding: 16px 32px; background: #000000; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px; transition: all 0.2s ease; border: none; cursor: pointer;">Accept Invitation</a>
          </div>

          <hr style="border: 0; height: 1px; background: #e5e5e5; margin: 60px 0;" />

          <p style="font-size: 17px; color: #626262; line-height: 1.8; margin-bottom: 40px; margin-top: 0;">
            If the button doesn't work, copy and paste this link into your browser:
          </p>

          <div style="background: #f5f5f5; padding: 24px; margin: 40px 0;">
            <span style="font-size: 12px; color: #626262; text-transform: uppercase; letter-spacing: 0.8px; margin-bottom: 12px; display: block; font-weight: 600;">Invitation Link</span>
            <a href="%s" style="color: #0066cc; text-decoration: none; font-size: 14px; word-break: break-all; line-height: 1.6;">%s</a>
          </div>

          <p style="font-size: 13px; color: #626262; line-height: 1.7; margin-top: 40px;">
            This invitation expires in 7 days. If you didn't expect this email, feel free to disregard it.
          </p>
        </div>

        <!-- Footer -->
        <div style="padding: 40px 0 0 0; text-align: left;">
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin-bottom: 8px; margin-top: 0;">© 2026 Modeva. All rights reserved.</p>
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin-top: 0;">
            Questions?
            <a href="mailto:support@modeva.shop" style="color: #0066cc; text-decoration: none; font-size: 13px; font-weight: 500;">Contact support</a>
          </p>
        </div>
      </div>
    </div>
  </body>
</html>`, data.AdminName, data.InviteLink, data.InviteLink, data.InviteLink)
}
