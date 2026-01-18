package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// BrowserScanner handles headless browser operations using Playwright.
type BrowserScanner struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	cfg     ScannerConfig
}

// ScannerConfig holds the configuration for the browser scanner.
type ScannerConfig struct {
	Headless   bool
	TimeoutMs  float64
	UseStealth bool
}

// NewBrowserScanner initializes a new Playwright scanner with config.
func NewBrowserScanner(cfg ScannerConfig) (*BrowserScanner, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(cfg.Headless),
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}
	return &BrowserScanner{
		pw:      pw,
		browser: browser,
		cfg:     cfg,
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
// Uses blocking to prevent loading images, fonts, and stylesheets.
func (b *BrowserScanner) ScanURL(ctx context.Context, url string) (string, []byte, error) {
	// 1. Stealth: Randomize User-Agent
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
	}
	// Simple random selection (pseudo-random is fine for MVP)
	// For better randomness, we'd use math/rand, but let's just pick one for now or loop used time.
	// Using a fixed one for stability in this snippet, but ideally random.
	ua := userAgents[time.Now().UnixNano()%int64(len(userAgents))]

	page, err := b.browser.NewPage(playwright.BrowserNewPageOptions{
		UserAgent: playwright.String(ua),
	})
	if err != nil {
		return "", nil, fmt.Errorf("could not create page: %w", err)
	}
	defer page.Close()

	// 2. Stealth: Inject Evasion Scripts (If Enabled)
	if b.cfg.UseStealth {
		// Mask navigator.webdriver
		if err := page.AddInitScript(playwright.Script{
			Content: playwright.String(`
                Object.defineProperty(navigator, 'webdriver', {
                    get: () => undefined,
                });
    
                // --- WebGL Spoofing (Stealth Module) ---
                // Override getParameter to hide Headless evidence
                const getParameter = WebGLRenderingContext.prototype.getParameter;
                WebGLRenderingContext.prototype.getParameter = function(parameter) {
                    // UNMASKED_VENDOR_WEBGL
                    if (parameter === 37445) {
                        return 'Intel Inc.';
                    }
                    // UNMASKED_RENDERER_WEBGL
                    if (parameter === 37446) {
                        return 'Intel Iris OpenGL Engine';
                    }
                    return getParameter(parameter);
                };
    
                // Pass basic bot tests
                window.chrome = { runtime: {} };
                Object.defineProperty(navigator, 'plugins', {
                    get: () => [1, 2, 3, 4, 5],
                });
                Object.defineProperty(navigator, 'languages', {
                    get: () => ['en-US', 'en'],
                });
            `),
		}); err != nil {
			// Log error but continue
			fmt.Printf("Warning: could not inject stealth scripts: %v\n", err)
		}
	}

	// 3. Enable Route Blocking for resources
	err = page.Route("**/*", func(route playwright.Route) {
		req := route.Request()
		resourceType := req.ResourceType()
		if resourceType == "image" || resourceType == "font" || resourceType == "stylesheet" || resourceType == "media" {
			route.Abort()
			return
		}
		route.Continue()
	})
	if err != nil {
		// Log error but continue?
	}

	// 2. Navigate
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}

	_, err = page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(b.cfg.TimeoutMs),
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
