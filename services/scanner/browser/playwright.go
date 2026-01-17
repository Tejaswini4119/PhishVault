package browser

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

// BrowserScanner handles headless browser operations using Playwright.
type BrowserScanner struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

// NewBrowserScanner initializes a new Playwright scanner.
// It launches the browser instance.
func NewBrowserScanner() (*BrowserScanner, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}
	return &BrowserScanner{
		pw:      pw,
		browser: browser,
	}, nil
}

// Close cleans up Playwright resources.
func (b *BrowserScanner) Close() error {
	if err := b.browser.Close(); err != nil {
		return err
	}
	return b.pw.Stop()
}

// ScanURL visits a URL and captures the DOM and screenshot.
// Returns HTML content, screenshot bytes, and error.
func (b *BrowserScanner) ScanURL(url string) (string, []byte, error) {
	page, err := b.browser.NewPage(playwright.BrowserNewPageOptions{
		UserAgent: playwright.String("PhishVault-Scanner/1.0"),
	})
	if err != nil {
		return "", nil, fmt.Errorf("could not create page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL with a timeout
	_, err = page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000), // 30 seconds
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return "", nil, fmt.Errorf("could not goto url: %w", err)
	}

	// Capture DOM
	content, err := page.Content()
	if err != nil {
		return "", nil, fmt.Errorf("could not get page content: %w", err)
	}

	// Capture Screenshot
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
		Type:     playwright.ScreenshotTypePng,
	})
	if err != nil {
		return content, nil, fmt.Errorf("could not take screenshot: %w", err)
	}

	return content, screenshot, nil
}
