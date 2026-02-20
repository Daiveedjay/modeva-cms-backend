package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// OrderInvoicePDFEmailData holds data for order invoice PDF email
type OrderInvoicePDFEmailData struct {
	CustomerName  string
	CustomerEmail string
	OrderNumber   string
	OrderDate     string
	DueDate       string
	AddressStreet string
	AddressCity   string
	AddressState  string
	AddressZip    string
	Items         []OrderInvoiceItem
	SubtotalTotal float64
	ShippingCost  float64
	Tax           float64
	Discount      float64
	TotalAmount   float64
	PDFContent    []byte
}

// OrderInvoiceItem represents a line item in an invoice
type OrderInvoiceItem struct {
	ProductName string
	Quantity    int
	Price       float64
	Subtotal    float64
}

// SendOrderInvoicePDFEmail sends an order invoice with HTML preview + PDF attachment via Resend
func (r *ResendClient) SendOrderInvoicePDFEmail(data OrderInvoicePDFEmailData) error {
	// Build invoice items HTML rows
	var itemsRows strings.Builder
	for _, item := range data.Items {
		itemsRows.WriteString(fmt.Sprintf(`
      <tr>
        <td style="padding: 8px 0; font-size: 14px; color: #262622;">%s</td>
        <td style="padding: 8px 0; font-size: 14px; text-align: right; color: #262622;">%d</td>
        <td style="padding: 8px 0; font-size: 14px; text-align: right; color: #262622;">$%.2f</td>
        <td style="padding: 8px 0; font-size: 14px; text-align: right; font-weight: 600; color: #262622;">$%.2f</td>
      </tr>
    `, item.ProductName, item.Quantity, item.Price, item.Subtotal))
	}

	// Discount row
	discountRow := ""
	if data.Discount > 0 {
		discountRow = fmt.Sprintf(`
    <tr>
      <td colspan="3" style="padding: 6px 0; font-size: 14px; color: #79776d;">Discount</td>
      <td style="padding: 6px 0; font-size: 14px; text-align: right; color: #262622;">-$%.2f</td>
    </tr>
    `, data.Discount)
	}

	// Address line
	addressLine := ""
	if data.AddressStreet != "" {
		addressLine = fmt.Sprintf("%s, %s %s %s",
			data.AddressStreet, data.AddressCity, data.AddressState, data.AddressZip)
	}

	// Build full HTML with inline styles
	var html strings.Builder
	html.WriteString(fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Invoice - %s</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif; background-color: #fafaf7; line-height: 1.5; padding: 16px;">
  <table width="100%%" cellpadding="0" cellspacing="0" border="0" style="max-width: 900px; margin: auto; background: #ffffff; padding: 24px;">
    <tr>
      <td style="border-bottom: 1px solid #e5e5e0; padding-bottom: 16px;">
        <h1 style="margin: 0; font-size: 30px; font-weight: bold; color: #262622;">INVOICE</h1>
      </td>
    </tr>

    <tr>
      <td style="padding: 16px 0;">
        <h2 style="margin: 0; font-size: 24px; font-weight: bold; color: #262622;">MODEVA STORE</h2>
        <p style="margin: 4px 0; font-size: 14px; color: #79776d;">contact@modeva.com</p>
      </td>
    </tr>

    <tr>
      <td style="padding: 16px 0;">
        <table width="100%%" cellpadding="0" cellspacing="0" border="0">
          <tr>
            <td style="vertical-align: top;">
              <p style="margin: 0; font-size: 14px; font-weight: bold; color: #262622;">Bill To</p>
              <p style="margin: 4px 0; font-size: 14px; color: #262622;">%s</p>
              <p style="margin: 4px 0; font-size: 14px; color: #79776d;">%s</p>
              <p style="margin: 4px 0; font-size: 14px; color: #79776d;">%s</p>
            </td>
            <td style="text-align: right; vertical-align: top;">
              <p style="margin: 0; font-size: 14px; color: #79776d;">Invoice Number</p>
              <p style="margin: 4px 0; font-size: 14px; font-weight: bold; color: #262622;">%s</p>
              <p style="margin: 8px 0 0 0; font-size: 14px; color: #79776d;">Invoice Date</p>
              <p style="margin: 4px 0; font-size: 14px; font-weight: bold; color: #262622;">%s</p>
              <p style="margin: 8px 0 0 0; font-size: 14px; color: #79776d;">Due Date</p>
              <p style="margin: 4px 0; font-size: 14px; font-weight: bold; color: #262622;">%s</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>

    <tr>
      <td style="padding: 16px 0; border-top: 1px solid #e5e5e0; border-bottom: 1px solid #e5e5e0;">
        <table width="100%%" cellpadding="0" cellspacing="0" border="0">
          <thead>
            <tr>
              <th style="text-align: left; font-size: 12px; text-transform: uppercase; color: #262622; padding-bottom: 8px;">Description</th>
              <th style="text-align: right; font-size: 12px; text-transform: uppercase; color: #262622; padding-bottom: 8px;">Qty</th>
              <th style="text-align: right; font-size: 12px; text-transform: uppercase; color: #262622; padding-bottom: 8px;">Price</th>
              <th style="text-align: right; font-size: 12px; text-transform: uppercase; color: #262622; padding-bottom: 8px;">Total</th>
            </tr>
          </thead>
          <tbody>
            %s
          </tbody>
        </table>
      </td>
    </tr>

    <tr>
      <td style="padding: 16px 0;">
        <table align="right" width="300" cellpadding="0" cellspacing="0" border="0">
          <tr>
            <td style="font-size: 14px; color: #79776d;">Subtotal</td>
            <td style="text-align: right; font-size: 14px; color: #262622;">$%.2f</td>
          </tr>
          <tr>
            <td style="font-size: 14px; color: #79776d;">Shipping</td>
            <td style="text-align: right; font-size: 14px; color: #262622;">$%.2f</td>
          </tr>
          <tr>
            <td style="font-size: 14px; color: #79776d;">Tax</td>
            <td style="text-align: right; font-size: 14px; color: #262622;">$%.2f</td>
          </tr>
          %s
          <tr>
            <td style="font-size: 14px; font-weight: bold; border-top: 1px solid #e5e5e0; padding-top: 8px;">Total</td>
            <td style="text-align: right; font-size: 16px; font-weight: bold; color: #262622; border-top: 1px solid #e5e5e0; padding-top: 8px;">$%.2f</td>
          </tr>
        </table>
      </td>
    </tr>

    <tr>
      <td style="padding: 16px 0; border-top: 1px solid #e5e5e0;">
        <p style="font-size: 14px; font-weight: bold; color: #262622;">Thank you for your business!</p>
        <p style="font-size: 14px; color: #79776d;">© 2026 Modeva Store. All rights reserved.</p>
      </td>
    </tr>

  </table>
</body>
</html>
`, data.OrderNumber,
		data.CustomerName, data.CustomerEmail, addressLine,
		data.OrderNumber, data.OrderDate, data.DueDate,
		itemsRows.String(),
		data.SubtotalTotal, data.ShippingCost, data.Tax,
		discountRow,
		data.TotalAmount,
	))

	htmlBody := html.String()

	// Encode PDF to base64
	pdfBase64 := base64.StdEncoding.EncodeToString(data.PDFContent)

	payload := map[string]interface{}{
		"from":    r.from,
		"to":      data.CustomerEmail,
		"subject": fmt.Sprintf("Your Invoice #%s from Modeva Store", data.OrderNumber),
		"html":    htmlBody,
		"attachments": []map[string]interface{}{
			{
				"filename": fmt.Sprintf("invoice-%s.pdf", data.OrderNumber),
				"content":  pdfBase64,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[resend] failed to marshal payload: %v", err)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("[resend] api returned status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("resend api error: status %d", resp.StatusCode)
	}

	log.Printf("[resend] order invoice email sent to %s for order %s", data.CustomerEmail, data.OrderNumber)
	return nil
}
