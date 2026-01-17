package parser

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
)

// EmailData holds extracted email information
type EmailData struct {
	Subject string
	From    string
	To      string
	Date    string
	Body    string // Plain text or HTML
	Headers map[string]string
}

// ParseEmail parses a raw email byte stream (e.g. from .eml)
func ParseEmail(r io.Reader) (*EmailData, error) {
	m, err := mail.ReadMessage(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	header := m.Header
	data := &EmailData{
		Subject: header.Get("Subject"),
		From:    header.Get("From"),
		To:      header.Get("To"),
		Date:    header.Get("Date"),
		Headers: make(map[string]string),
	}

	// Store key interesting headers for analysis
	interestingHeaders := []string{"Received", "DKIM-Signature", "Received-SPF", "Authentication-Results"}
	for _, h := range interestingHeaders {
		if v := header.Get(h); v != "" {
			data.Headers[h] = v
		}
	}

	// Basic body extraction (Not robust for multipart, meant for MVP text/plain)
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(m.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	data.Body = buf.String()

	return data, nil
}

// AnalyzeSPF delegates to the robust AuthHeader parser.
// In the future, this function signature might change to return the full AuthResult,
// but for compatibility with main.go we return a summary string for now.
func AnalyzeSPF(headers map[string]string) string {
	// Use the robust parser
	res := AnalyzeAuthHeaders(headers)
	return res.SPF
}
