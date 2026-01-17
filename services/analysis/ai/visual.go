package ai

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/corona10/goimagehash"
)

// CompareScreenshots calculates the similarity between two images using Perceptual Hashing (dHash).
// output: 0.0 (no match) to 1.0 (perfect match).
func CompareScreenshots(targetImgData, goldenImgData []byte) (float64, error) {
	// Decode images
	targetImg, _, err := image.Decode(bytes.NewReader(targetImgData))
	if err != nil {
		return 0, err
	}
	goldenImg, _, err := image.Decode(bytes.NewReader(goldenImgData))
	if err != nil {
		return 0, err
	}

	// Calculate dHash ( Difference Hash) for both
	// dHash is fast and robust to scaling/color changes.
	targetHash, err := goimagehash.DifferenceHash(targetImg)
	if err != nil {
		return 0, err
	}
	goldenHash, err := goimagehash.DifferenceHash(goldenImg)
	if err != nil {
		return 0, err
	}

	// Compute Hamming Distance
	distance, err := targetHash.Distance(goldenHash)
	if err != nil {
		return 0, err
	}

	// Normalize distance to a similarity score (0.0 to 1.0)
	// dHash is usually 64 bits. Distance 0 = Identical. Distance 64 = Opposite.
	// Similarity = 1 - (distance / 64)
	score := 1.0 - (float64(distance) / 64.0)
	if score < 0 {
		score = 0
	}
	return score, nil
}
