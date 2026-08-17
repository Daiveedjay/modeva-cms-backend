package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// CustomerUnbanEmailData holds data for customer unban notification
type CustomerUnbanEmailData struct {
	CustomerName  string
	CustomerEmail string
	UnbanReason   string
	SupportEmail  string
}

// SendCustomerUnbanEmail sends an unban notification email to a customer
func (r *ResendClient) SendCustomerUnbanEmail(data CustomerUnbanEmailData) error {
	htmlBody := r.buildCustomerUnbanEmailHTML(data)

	payload := map[string]interface{}{
		"from":    r.from,
		"to":      data.CustomerEmail,
		"subject": "Account Reinstated - Modeva",
		"html":    htmlBody,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("resend api error: status %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("[resend] unban notification sent to %s", data.CustomerEmail)
	return nil
}

// buildCustomerUnbanEmailHTML creates HTML for unban notification
func (r *ResendClient) buildCustomerUnbanEmailHTML(data CustomerUnbanEmailData) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Account Reinstated</title>
  </head>
  <body style="margin: 0; padding: 0; box-sizing: border-box; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif; background-color: #ffffff; color: #1a1a1a; line-height: 1.6;">
    <div style="background-color: #ffffff; padding: 60px 20px;">
      <div style="max-width: 600px; margin: 0 auto; background: #ffffff;">
        <div style="padding: 0 0 80px 0; text-align: left;">
          <div style="font-size: 24px; font-weight: 700; color: #1a1a1a; letter-spacing: -0.3px;">Modeva</div>
        </div>
        <div style="padding: 0;">
          <p style="font-size: 36px; font-weight: 700; color: #16a34a; margin-bottom: 24px; letter-spacing: -0.8px; line-height: 1.2; margin-top: 0;">Account Reinstated</p>
          <p style="font-size: 17px; color: #626262; line-height: 1.8; margin-bottom: 40px; margin-top: 0;">
            Hi <span style="color: #000000; font-weight: 600;">%s</span>,
          </p>
          <p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px;">
            Great news! Your Modeva account suspension has been lifted and your account has been reinstated.
          </p>
          <div style="background: #f0fdf4; border-left: 4px solid #16a34a; padding: 24px; margin: 40px 0;">
            <p style="font-size: 15px; color: #166534; line-height: 1.6; margin: 0; font-weight: 500;">%s</p>
          </div>
          <p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px;">
            You now have full access to your account and can resume shopping. We appreciate your understanding during this time.
          </p>
          <div style="text-align: left; margin: 50px 0 60px 0;">
            <a href="https://modeva.biz" style="display: inline-block; padding: 16px 32px; background: #000000; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px; transition: all 0.2s ease;">Continue Shopping</a>
          </div>
          <hr style="border: 0; height: 1px; background: #e5e5e5; margin: 60px 0;" />
          <p style="font-size: 13px; color: #626262; line-height: 1.7; margin-top: 40px;">
            If you have any questions, contact us at <a href="mailto:%s" style="color: #0066cc; text-decoration: none;">%s</a>
          </p>
        </div>
        <div style="padding: 40px 0 0 0; text-align: left;">
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin: 0;">© 2026 Modeva. All rights reserved.</p>
        </div>
      </div>
    </div>
  </body>
</html>`, data.CustomerName, data.UnbanReason, data.SupportEmail, data.SupportEmail)
}
