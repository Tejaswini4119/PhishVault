package utils

import (
	"net/http"
	"net/url"
	"time"
)

// CanonicalizeURL strips tracking parameters and resolves the final URL.
// It returns the cleaned URL or an error if processing fails.
func CanonicalizeURL(rawURL string) (string, error) {
	// 1. Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 2. Strip tracking parameters
	q := parsedURL.Query()
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "ref", "source",
	}
	for _, param := range trackingParams {
		q.Del(param)
	}
	parsedURL.RawQuery = q.Encode()
	cleanURL := parsedURL.String()

	// 3. Resolve redirects (Unwind)
	// Use a custom HTTP client to follow redirects and get the final URL.
	// We use a HEAD request first to minimize data transfer.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Head(cleanURL)
	if err != nil {
		// Fallback to GET if HEAD fails (some servers block HEAD)
		resp, err = client.Get(cleanURL)
		if err != nil {
			return cleanURL, err // Return cleaned URL if we can't resolve
		}
	}
	defer resp.Body.Close()

	// The final URL is in resp.Request.URL if redirects were followed
	finalURL := resp.Request.URL.String()

	// 4. Strip params again from the final URL just in case
	finalParsed, err := url.Parse(finalURL)
	if err == nil {
		q = finalParsed.Query()
		for _, param := range trackingParams {
			q.Del(param)
		}
		finalParsed.RawQuery = q.Encode()
		return finalParsed.String(), nil
	}

	return finalURL, nil
}
