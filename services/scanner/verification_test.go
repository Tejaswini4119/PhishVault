package scanner_test

import (
	"context"
	"testing"

	"github.com/PhishVault/PhishVault-2/services/scanner/browser"
	fetcher "github.com/PhishVault/PhishVault-2/services/scanner/http"
	"github.com/PhishVault/PhishVault-2/services/scanner/utils"
)

func TestCanonicalize(t *testing.T) {
	raw := "http://example.com/page?utm_source=twitter&ref=123"
	expected := "http://example.com/page"
	got, err := utils.CanonicalizeURL(raw)
	if err != nil {
		t.Fatalf("Canonicalize failed: %v", err)
	}
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
	t.Logf("Canonicalized: %s", got)
}

func TestFetch(t *testing.T) {
	f := fetcher.NewFetcher()
	code, headers, err := f.FetchHeaders(context.Background(), "http://example.com")

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	t.Logf("Status: %d", code)
	t.Logf("Headers: %v", headers)
}

func TestBrowser(t *testing.T) {
	// Skip verification if Playwright is not installed fully
	// This test might fail if 'playwright install' wasn't run.
	cfg := browser.ScannerConfig{
		Headless:   true,
		TimeoutMs:  30000,
		UseStealth: false,
	}
	b, err := browser.NewBrowserScanner(cfg)
	if err != nil {
		t.Skipf("Skipping browser test: %v", err)
	}
	defer b.Close()

	content, screenshot, err := b.ScanURL(context.Background(), "http://example.com")
	if err != nil {

		t.Fatalf("Browser scan failed: %v", err)
	}
	t.Logf("Content length: %d", len(content))
	t.Logf("Screenshot size: %d", len(screenshot))
}
