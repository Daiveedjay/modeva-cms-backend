package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// CustomerBanEmailData holds data for customer ban notification
type CustomerBanEmailData struct {
	CustomerName  string
	CustomerEmail string
	BanReason     string
	SupportEmail  string
}

// CustomerDeletionEmailData holds data for customer deletion notification
type CustomerDeletionEmailData struct {
	CustomerName  string
	CustomerEmail string
	SupportEmail  string
}

// SendCustomerEmail sends a custom email to a customer via Resend
// func (r *ResendClient) SendCustomerEmail(data CustomerEmailData) error {
// 	htmlBody := r.buildCustomerEmailHTML(data)

// 	payload := map[string]interface{}{
// 		"from":    r.from,
// 		"to":      data.CustomerEmail,
// 		"subject": data.Subject,
// 		"html":    htmlBody,
// 	}

// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal payload: %w", err)
// 	}

// 	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
// 	if err != nil {
// 		return fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("failed to send request: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, _ := io.ReadAll(resp.Body)
// 	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
// 		return fmt.Errorf("resend api error: status %d - %s", resp.StatusCode, string(body))
// 	}

// 	log.Printf("[resend] customer email sent to %s", data.CustomerEmail)
// 	return nil
// }

// SendCustomerBanEmail sends a ban notification email to a customer
func (r *ResendClient) SendCustomerBanEmail(data CustomerBanEmailData) error {
	htmlBody := r.buildCustomerBanEmailHTML(data)

	payload := map[string]interface{}{
		"from":    r.from,
		"to":      data.CustomerEmail,
		"subject": "Account Suspended - Modeva",
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

	log.Printf("[resend] ban notification sent to %s", data.CustomerEmail)
	return nil
}

// SendCustomerDeletionEmail sends account deletion confirmation to a customer
func (r *ResendClient) SendCustomerDeletionEmail(data CustomerDeletionEmailData) error {
	htmlBody := r.buildCustomerDeletionEmailHTML(data)

	payload := map[string]interface{}{
		"from":    r.from,
		"to":      data.CustomerEmail,
		"subject": "Account Deleted - Modeva",
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

	log.Printf("[resend] deletion confirmation sent to %s", data.CustomerEmail)
	return nil
}

// buildCustomerBanEmailHTML creates HTML for ban notification
func (r *ResendClient) buildCustomerBanEmailHTML(data CustomerBanEmailData) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Account Suspended</title>
  </head>
  <body style="margin: 0; padding: 0; box-sizing: border-box; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif; background-color: #ffffff; color: #1a1a1a; line-height: 1.6;">
    <div style="background-color: #ffffff; padding: 60px 20px;">
      <div style="max-width: 600px; margin: 0 auto; background: #ffffff;">
        <div style="padding: 0 0 80px 0; text-align: left;">
          <div style="font-size: 24px; font-weight: 700; color: #1a1a1a; letter-spacing: -0.3px;">Modeva</div>
        </div>
        <div style="padding: 0;">
          <p style="font-size: 36px; font-weight: 700; color: #dc2626; margin-bottom: 24px; letter-spacing: -0.8px; line-height: 1.2; margin-top: 0;">Account Suspended</p>
          <p style="font-size: 17px; color: #626262; line-height: 1.8; margin-bottom: 40px; margin-top: 0;">
            Hi <span style="color: #000000; font-weight: 600;">%s</span>,
          </p>
          <p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px;">
            Your Modeva account has been suspended:
          </p>
          <div style="background: #fef2f2; border-left: 4px solid #dc2626; padding: 24px; margin: 40px 0;">
            <p style="font-size: 15px; color: #991b1b; line-height: 1.6; margin: 0; font-weight: 500;">%s</p>
          </div>
          <p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px;">
            You cannot access your account during this suspension. Your data remains secure.
          </p>
          <hr style="border: 0; height: 1px; background: #e5e5e5; margin: 60px 0;" />
          <div style="background: #f5f5f5; padding: 24px; margin: 40px 0;">
            <p style="font-size: 15px; color: #1a1a1a; line-height: 1.6; margin: 0;">
              <strong>Need to appeal?</strong><br/>
              Contact <a href="mailto:%s" style="color: #0066cc; text-decoration: none;">%s</a>
            </p>
          </div>
        </div>
        <div style="padding: 40px 0 0 0; text-align: left;">
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin: 0;">© 2026 Modeva. All rights reserved.</p>
        </div>
      </div>
    </div>
  </body>
</html>`, data.CustomerName, data.BanReason, data.SupportEmail, data.SupportEmail)
}

// buildCustomerDeletionEmailHTML creates HTML for deletion confirmation
func (r *ResendClient) buildCustomerDeletionEmailHTML(data CustomerDeletionEmailData) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Account Deleted</title>
  </head>
  <body style="margin: 0; padding: 0; box-sizing: border-box; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif; background-color: #ffffff; color: #1a1a1a; line-height: 1.6;">
    <div style="background-color: #ffffff; padding: 60px 20px;">
      <div style="max-width: 600px; margin: 0 auto; background: #ffffff;">
        <div style="padding: 0 0 80px 0; text-align: left;">
          <div style="font-size: 24px; font-weight: 700; color: #1a1a1a; letter-spacing: -0.3px;">Modeva</div>
        </div>
        <div style="padding: 0;">
          <p style="font-size: 36px; font-weight: 700; color: #000000; margin-bottom: 24px; letter-spacing: -0.8px; line-height: 1.2; margin-top: 0;">Account Deleted</p>
          <p style="font-size: 17px; color: #626262; line-height: 1.8; margin-bottom: 40px; margin-top: 0;">
            Hi <span style="color: #000000; font-weight: 600;">%s</span>,
          </p>
          <p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px;">
            Your Modeva account has been permanently deleted.
          </p>
          <div style="background: #fef3c7; border-left: 4px solid #f59e0b; padding: 24px; margin: 40px 0;">
            <p style="font-size: 15px; color: #92400e; line-height: 1.6; margin: 0; font-weight: 500;">
              <strong>What happens next?</strong><br/>
              • All personal data has been removed<br/>
              • Order history is no longer accessible<br/>
              • No more marketing communications<br/>
              • This action cannot be undone
            </p>
          </div>
          <p style="font-size: 17px; color: #1a1a1a; line-height: 1.8; margin-bottom: 24px;">
            Thank you for being part of Modeva. You're always welcome to return.
          </p>
          <hr style="border: 0; height: 1px; background: #e5e5e5; margin: 60px 0;" />
          <p style="font-size: 13px; color: #626262; line-height: 1.7; margin-top: 40px;">
            Questions? Contact <a href="mailto:%s" style="color: #0066cc; text-decoration: none;">%s</a>
          </p>
        </div>
        <div style="padding: 40px 0 0 0; text-align: left;">
          <p style="font-size: 13px; color: #626262; line-height: 1.8; margin: 0;">© 2026 Modeva. All rights reserved.</p>
        </div>
      </div>
    </div>
  </body>
</html>`, data.CustomerName, data.SupportEmail, data.SupportEmail)
}
