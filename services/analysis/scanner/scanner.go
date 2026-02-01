package scanner

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/playwright-community/playwright-go"
)

type Scanner struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

func NewScanner() (*Scanner, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	return &Scanner{
		pw:      pw,
		browser: browser,
	}, nil
}

func (s *Scanner) Scan(ctx context.Context, targetURL string) (domain.Artifacts, error) {
	page, err := s.browser.NewPage()
	if err != nil {
		return domain.Artifacts{}, fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	// Timeout context
	// Playwright has its own timeout options, but we can respect context cancellation too.

	log.Printf("Scanning URL: %s", targetURL)
	if _, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000), // 30s
	}); err != nil {
		return domain.Artifacts{}, fmt.Errorf("could not goto url: %v", err)
	}

	// Get Content
	content, err := page.Content()
	if err != nil {
		return domain.Artifacts{}, fmt.Errorf("could not get content: %v", err)
	}

	// Screenshot
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type: playwright.ScreenshotTypePng,
	})
	if err != nil {
		return domain.Artifacts{}, fmt.Errorf("could not take screenshot: %v", err)
	}

	// Mock Visual Hash (in prod this would use image hashing lib)
	// For now, return a dummy hash or hash of the screenshot bytes
	// visualHash := fmt.Sprintf("%x", sha256.Sum256(screenshot))
	// But ProcessArtifact expects uint64 hex string?
	// Let's just hardcode a non-zero hash to trigger analysis if needed,
	// or rely on NLP.
	// If we want "Brand Match", we need a known hash.
	// Let's leave it empty or random for now unless we have image hashing.
	visualHash := "1234567890abcdef"

	return domain.Artifacts{
		RawContent: content,
		Screenshot: base64.StdEncoding.EncodeToString(screenshot), // Store base64 or path
		VisualHash: visualHash,
	}, nil
}

func (s *Scanner) Close() {
	if s.browser != nil {
		s.browser.Close()
	}
	if s.pw != nil {
		s.pw.Stop()
	}
}
