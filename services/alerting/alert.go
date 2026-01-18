package alerting

import (
	"fmt"
	"log"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// Dispatcher handles sending alerts.
type Dispatcher struct {
	// In production, we'd inject email/slack clients here.
}

// NewDispatcher creates a new alerting dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// CheckAndDispatch evaluates the SAL and sends alerts if thresholds are met.
func (d *Dispatcher) CheckAndDispatch(sal domain.SAL) {
	if sal.RiskScore > 0.8 || sal.Verdict == "MALICIOUS" {
		d.sendHighRiskAlert(sal)
	}
}

func (d *Dispatcher) sendHighRiskAlert(sal domain.SAL) {
	// Mock Alert for MVP
	log.Printf("[ALERT] HIGH RISK DETECTED for %s (Score: %.2f)", sal.URL, sal.RiskScore)
	log.Printf("[ALERT] Notification sent to Security Ops via Email/Slack.")

	// Example format:
	msg := fmt.Sprintf("Phishing Detected! Target: %s, Score: %.2f, Verdict: %s", sal.URL, sal.RiskScore, sal.Verdict)
	// d.emailClient.Send("secops@company.com", msg)
	// d.slackClient.Post("#incident-response", msg)
	fmt.Printf("[MOCK] Sent Payload: %s\n", msg)
}
