package notifier

import (
	"Monitra/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// Notifier handles sending notifications
type Notifier struct {
	config             models.NotificationConfig
	previousState      map[string]bool // tracks previous up/down state
	sslNotifiedTargets map[string]bool // tracks if SSL expiring notification was sent
}

// New creates a new notifier instance
func New(config models.NotificationConfig) *Notifier {
	return &Notifier{
		config:             config,
		previousState:      make(map[string]bool),
		sslNotifiedTargets: make(map[string]bool),
	}
}

// Notify sends notifications based on check results
func (n *Notifier) Notify(result *models.CheckResult) {
	// Check if state changed
	previouslyUp, exists := n.previousState[result.TargetName]
	n.previousState[result.TargetName] = result.IsUp

	var event string
	if !exists {
		// First check, don't notify
		return
	}

	if result.IsUp && !previouslyUp {
		event = "up"
		// Reset SSL notification flag when service comes back up
		delete(n.sslNotifiedTargets, result.TargetName)
	} else if !result.IsUp && previouslyUp {
		event = "down"
	} else if result.IsUp && result.SSLDaysLeft != nil && *result.SSLDaysLeft <= 30 {
		// Only notify once for SSL expiring
		if n.sslNotifiedTargets[result.TargetName] {
			return
		}
		event = "ssl_expiring"
		n.sslNotifiedTargets[result.TargetName] = true
	} else {
		// No notification needed
		return
	}

	// Send webhook notifications
	for _, webhook := range n.config.Webhooks {
		if webhook.Enabled && n.shouldNotify(webhook.Events, event) {
			go n.sendWebhook(webhook.URL, result, event)
		}
	}

	// Send email notifications
	if n.config.Email.Enabled && n.shouldNotify(n.config.Email.Events, event) {
		go n.sendEmail(result, event)
	}
}

// shouldNotify checks if an event should trigger notification
func (n *Notifier) shouldNotify(events []string, event string) bool {
	if len(events) == 0 {
		return true // notify on all events if none specified
	}
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}

// sendWebhook sends a webhook notification
func (n *Notifier) sendWebhook(url string, result *models.CheckResult, event string) {
	payload := map[string]interface{}{
		"event":         event,
		"target_name":   result.TargetName,
		"target_url":    result.TargetURL,
		"status_code":   result.StatusCode,
		"response_time": result.ResponseTime,
		"is_up":         result.IsUp,
		"error":         result.Error,
		"checked_at":    result.CheckedAt.Format(time.RFC3339),
	}

	if result.SSLDaysLeft != nil {
		payload["ssl_days_left"] = *result.SSLDaysLeft
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send webhook to %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("Webhook sent successfully to %s\n", url)
	} else {
		fmt.Printf("Webhook failed with status %d\n", resp.StatusCode)
	}
}

// sendEmail sends an email notification
func (n *Notifier) sendEmail(result *models.CheckResult, event string) {
	if n.config.Email.SMTP.Host == "" {
		return
	}

	subject := fmt.Sprintf("[Monitra] %s - %s", strings.ToUpper(event), result.TargetName)
	body := n.buildEmailBody(result, event)

	msg := []byte(fmt.Sprintf("Subject: %s\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", subject, n.config.Email.From, strings.Join(n.config.Email.To, ","), body))

	auth := smtp.PlainAuth("",
		n.config.Email.SMTP.Username,
		n.config.Email.SMTP.Password,
		n.config.Email.SMTP.Host,
	)

	addr := fmt.Sprintf("%s:%d", n.config.Email.SMTP.Host, n.config.Email.SMTP.Port)
	err := smtp.SendMail(addr, auth, n.config.Email.From, n.config.Email.To, msg)
	if err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
		return
	}

	fmt.Printf("Email notification sent successfully\n")
}

// buildEmailBody creates the email body content
func (n *Notifier) buildEmailBody(result *models.CheckResult, event string) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("Event: %s\n", strings.ToUpper(event)))
	body.WriteString(fmt.Sprintf("Target: %s\n", result.TargetName))
	body.WriteString(fmt.Sprintf("URL: %s\n", result.TargetURL))
	body.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "UP", false: "DOWN"}[result.IsUp]))
	body.WriteString(fmt.Sprintf("Status Code: %d\n", result.StatusCode))
	body.WriteString(fmt.Sprintf("Response Time: %dms\n", result.ResponseTime))

	if result.Error != "" {
		body.WriteString(fmt.Sprintf("Error: %s\n", result.Error))
	}

	if result.SSLDaysLeft != nil {
		body.WriteString(fmt.Sprintf("SSL Days Left: %d\n", *result.SSLDaysLeft))
	}

	body.WriteString(fmt.Sprintf("Checked At: %s\n", result.CheckedAt.Format(time.RFC3339)))

	return body.String()
}
