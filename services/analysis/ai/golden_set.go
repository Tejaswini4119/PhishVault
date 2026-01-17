package ai

// GoldenSet matches Feature 5.9: "Golden Set" of known brands.
// Maps Brand Name -> List of Known Visual Hashes (pHash/dHash).
type GoldenSet struct {
	Brands map[string][]uint64
}

var GlobalGoldenSet *GoldenSet

func InitGoldenSet() {
	GlobalGoldenSet = &GoldenSet{
		Brands: make(map[string][]uint64),
	}

	// Load known hashes (Simulated for MVP)
	// In production, this loads from DB/JSON.

	// Example: Microsoft Login Page Hash (Hypothetical)
	GlobalGoldenSet.Add("Microsoft", 0x1234567890ABCDEF)
	// Example: PayPal Login
	GlobalGoldenSet.Add("PayPal", 0x9876543210ABCDEF)
	// Example: Google Login
	GlobalGoldenSet.Add("Google", 0xAAAA5555AAAA5555)
}

func (gs *GoldenSet) Add(brand string, hash uint64) {
	gs.Brands[brand] = append(gs.Brands[brand], hash)
}

// FindMatch checks if the provided hash matches any brand in the Golden Set.
// Returns: BrandName, Similarity (0.0 - 1.0), IsMatch
func (gs *GoldenSet) FindMatch(inputHash uint64) (string, float64, bool) {
	bestBrand := ""
	bestScore := 0.0

	for brand, hashes := range gs.Brands {
		for _, h := range hashes {
			dist := hamming(inputHash, h)
			// Normalized Similarity: 1.0 - (dist / 64)
			score := 1.0 - (float64(dist) / 64.0)

			if score > bestScore {
				bestScore = score
				bestBrand = brand
			}
		}
	}

	// Threshold 0.85 per Technical Workbook Section 5.9
	if bestScore > 0.85 {
		return bestBrand, bestScore, true
	}

	return "", bestScore, false
}

// Helper (Duplicated from dbscan for independence, or moved to utils in refactor)
func hamming(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	dist := 0
	for xor > 0 {
		if xor&1 == 1 {
			dist++
		}
		xor >>= 1
	}
	return dist
}
