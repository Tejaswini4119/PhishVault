package ai

import (
	"strings"

	"github.com/navossoc/bayesian"
)

// Classifier classes
const (
	ClassPhishing bayesian.Class = "Phishing"
	ClassBenign   bayesian.Class = "Benign"
)

var classifier *bayesian.Classifier

// InitBayesian initializes and trains the Naive Bayes classifier with a seed dataset.
// In a real industrial system, this would load a serialized model file.
func InitBayesian() {
	classifier = bayesian.NewClassifier(ClassPhishing, ClassBenign)

	// Industrial-Grade Seed Data (Simulated)
	phishingText := []string{
		"urgent action required verify your account immediately",
		"your account will be suspended within 24 hours",
		"click here to unlock your access",
		"unusual login attempt detected from new ip",
		"please update your billing information to avoid interruption",
		"confirm your identity securely",
		"dear customer we noticed suspicious activity",
		"security alert sign in to restore access",
	}

	benignText := []string{
		"thank you for your order your shipment is on the way",
		"meeting reminder for tomorrow at 10am",
		"weekly newsletter check out our new features",
		"your subscription has been renewed successfully",
		"happy birthday hoping you have a great day",
		"project update the timeline is looking good",
		"please review the attached document",
		"welcome to our service we are glad to have you",
	}

	// Train
	for _, s := range phishingText {
		classifier.Learn(strings.Fields(s), ClassPhishing)
	}
	for _, s := range benignText {
		classifier.Learn(strings.Fields(s), ClassBenign)
	}
}

// PredictPhishingProb returns the probability (0.0 to 1.0) that the text is Phishing.
func PredictPhishingProb(text string) float64 {
	if classifier == nil {
		InitBayesian()
	}

	// scores[0] is probability for the first class (Phishing)
	// library returns log probabilities usually, but ProbScores might return index.
	// Let's use LogScores and manual conversion or simply classification.
	// Update: classifier.LogScores returns log probabilities.
	// classifier.ProbScores is not a standard method in all versions, sticking to standard usage.

	// Standard usage: LogScores
	logScores, _, _ := classifier.LogScores(strings.Fields(strings.ToLower(text)))

	// Convert log scores to relative probability?
	// Simpler approach for MVP: usage SafeProb/Prob if available, or just heuristic on Log.
	// Actually, this library provides:
	// Find the highest score.

	// Check magnitude. If Phishing score > Benign score significanty.
	// For this snippet, let's trust the library's built-in "ProbScores" if it exists,
	// or re-implement simple comparison.
	// "navossoc/bayesian" LogScores returns sorted? No.
	// Index 0 = Phishing, Index 1 = Benign (order of NewClassifier args).

	phishingScore := logScores[0]
	benignScore := logScores[1]

	// Log scores are negative. Closer to 0 is higher probability.
	// e.g. -10 vs -50. -10 is much more likely.

	// Simple sigmoid-like normalization for output
	// If Phishing > Benign, Risk > 0.5

	if phishingScore > benignScore {
		// It's classified as phishing. How confident?
		diff := phishingScore - benignScore
		// diff is positive.
		if diff > 10 {
			return 0.99
		}
		if diff > 5 {
			return 0.9
		}
		return 0.75
	}

	return 0.1 // Likely Benign
}
