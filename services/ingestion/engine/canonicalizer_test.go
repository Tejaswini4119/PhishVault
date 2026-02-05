package engine

import (
	"context"
	"testing"
)

func TestCanonicalizerProcessor_Process(t *testing.T) {
	proc := NewCanonicalizerProcessor()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal URL",
			input:    "https://example.com/foo",
			expected: "https://example.com/foo",
		},
		{
			name:     "Strip Tracking Params",
			input:    "https://example.com/foo?utm_source=twitter&bar=baz&fbclid=123",
			expected: "https://example.com/foo?bar=baz",
		},
		{
			name:     "Punycode Decoding",
			input:    "http://xn--caf-dma.com",
			expected: "http://café.com",
		},
		{
			name:     "Scheme Normalization",
			input:    "HTTP://EXAMPLE.COM",
			expected: "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifact, err := proc.Process(ctx, []byte(tt.input), "test-source", nil)
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}
			if artifact.Content != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, artifact.Content)
			}
		})
	}
}
