package provider

import "time"

type WHOISData struct {
	Registrar    string
	CreationDate time.Time
	Country      string
	ASN          string
}

// FetchWHOIS simulates a WHOIS/pDNS lookup for a domain.
// In a real implementation, this would call an API like WhoisXML or query port 43.
func FetchWHOIS(domain string) (*WHOISData, error) {
	// Mock MVP Data
	return &WHOISData{
		Registrar:    "NameCheap, Inc.",
		CreationDate: time.Now().AddDate(-1, 0, 0), // 1 year old
		Country:      "US",
		ASN:          "AS13335 Cloudflare",
	}, nil
}
