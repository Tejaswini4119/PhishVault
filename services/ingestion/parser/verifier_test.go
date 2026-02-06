package parser

import (
	"strings"
	"testing"
)

func TestVerifyDKIM_NoSignature(t *testing.T) {
	// Simple email with no DKIM header
	email := "From: test@example.com\r\nSubject: Hello\r\n\r\nBody"
	r := strings.NewReader(email)

	result, err := VerifyDKIM(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "NONE" {
		t.Errorf("Expected result NONE, got %s", result)
	}
}

func TestVerifyDKIM_Malformed(t *testing.T) {
	// Testing coverage for error paths
	// Empty reader might return NONE or error depending on library
	r := strings.NewReader("")

	_, _ = VerifyDKIM(r)
	// We just want to ensure it doesn't panic
}
