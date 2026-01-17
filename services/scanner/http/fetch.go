package fetcher

import (
	"net/http"
	"time"
)

// Fetcher handles lightweight HTTP requests.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new instance of Fetcher with a default timeout.
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects for the raw fetcher, we want to see the immediate response.
				// Or do we? The plan says "Fetch headers, status codes... without rendering".
				// Usually scanners want to see the final page but lighter.
				// Let's follow redirects for now to be useful, or we can make it configurable.
				// Default behavior follows 10 redirects.
				return nil
			},
		},
	}
}

// FetchHeaders retrieves the headers and status code for a given URL.
// It performs a GET request but closes the body immediately to save bandwidth.
func (f *Fetcher) FetchHeaders(targetURL string) (int, http.Header, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return 0, nil, err
	}

	// Set a generic User-Agent to avoid some blocks
	req.Header.Set("User-Agent", "PhishVault-Scanner/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, resp.Header, nil
}
