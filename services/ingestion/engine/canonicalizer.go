package engine

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

type CanonicalizerProcessor struct {
	// Configurable options could go here (e.g., MaxRedirects)
}

func NewCanonicalizerProcessor() *CanonicalizerProcessor {
	return &CanonicalizerProcessor{}
}

func (p *CanonicalizerProcessor) Process(ctx context.Context, input []byte, sourceID string, metadata map[string]interface{}) (*IngestedArtifact, error) {
	rawURL := string(input)

	// 1. Basic Parse
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// 2. Punycode Transparency (ToUnicode)
	host, err := idna.ToUnicode(parsed.Host)
	if err == nil {
		parsed.Host = host
	}

	// 3. Remove Tracking Parameters
	q := parsed.Query()
	trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid", "_ga"}
	for _, param := range trackingParams {
		q.Del(param)
	}
	parsed.RawQuery = q.Encode()

	// 4. Redirect Unwinding (Headless/HTTP)
	// We'll use a custom HTTP client helper to determine the final URL.
	// This captures the redirect chain.
	finalURL, chain, err := p.unwindRedirects(parsed.String())
	if err != nil {
		// Log error but maybe proceed with the current URL if unwind fails?
		// For strictness, let's treat it as the canonical one or fail.
		// We'll keep the canonical as the best effort result.
		finalURL = parsed.String()
	}

	// 5. Final Normalization of the Result
	finalParsed, _ := url.Parse(finalURL)
	if finalParsed != nil {
		host, _ := idna.ToUnicode(finalParsed.Host)
		finalParsed.Host = strings.ToLower(host)
		finalParsed.Scheme = strings.ToLower(finalParsed.Scheme)
		finalParsed.Fragment = "" // Strip fragments usually

		// Force Unicode Host in output (net/url String() might Encode it)
		// We manually construct to satisfy "Punycode Transparency" requirement for internal storage
		finalURL = finalParsed.Scheme + "://" + finalParsed.Host
		if finalParsed.Path != "" || finalParsed.RawQuery != "" {
			// Use EscapedPath() if we want standard encoding for path, or just Path if we want raw?
			// Usually Path is decoded. EscapedPath is encoded.
			// Let's use RequestURI() equivalent but safer.
			// If Path is empty, we don't start with /. Parse logic usually ensures Path starts with / if present?
			// Actually URL.Path doesn't contain leading / for relative? But for absolute it does?
			// Let's rely on string concatenation carefully or use String() and replace Host?
			// Replacing Host in the String() output is tricky if it was encoded.
			// Let's append Path/Query.
			if !strings.HasPrefix(finalParsed.Path, "/") && finalParsed.Path != "" {
				finalURL += "/"
			}
			finalURL += finalParsed.EscapedPath()
			if finalParsed.RawQuery != "" {
				finalURL += "?" + finalParsed.RawQuery
			}
		}
	}

	// Construct Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["redirect_chain"] = chain
	metadata["original_url"] = rawURL

	return &IngestedArtifact{
		Type:            ArtifactTypeURL,
		Content:         finalURL,
		SourceID:        sourceID,
		Metadata:        metadata,
		Timestamp:       time.Now(),
		IngestionSource: "CANONICALIZER",
	}, nil
}

// unwindRedirects follows 3xx redirects up to a limit and returns the final URL and chain.
func (p *CanonicalizerProcessor) unwindRedirects(startURL string) (string, []string, error) {
	var chain []string
	chain = append(chain, startURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 12 {
				return http.ErrUseLastResponse
			}
			chain = append(chain, req.URL.String())
			return nil
		},
	}

	// We use HEAD first to be lighter, but GET is safer for some servers that block HEAD.
	// Using GET with a small range or just standard GET is robust.
	resp, err := client.Get(startURL)
	if err != nil {
		return startURL, chain, err
	}
	defer resp.Body.Close()

	// The final URL in the response struct is the one after redirects
	final := resp.Request.URL.String()
	if final != chain[len(chain)-1] {
		chain = append(chain, final)
	}

	return final, chain, nil
}
