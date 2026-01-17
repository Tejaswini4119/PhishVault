package phishvault.policy

default verdict = "SAFE"
default risk_score = 0.0

# Calculate Risk Score
risk_score = score {
	score := (visual_score * 0.5) + (urgency_score * 0.3) + (base_risk * 0.2)
}

# Factors
visual_score = input.visual_match_score
urgency_score = input.nlp_urgency_score

# Base Risk (e.g., from age or blacklists - placeholder for now)
base_risk = 1.0 { input.domain_age_days < 30 } else = 0.0

# Verdict Rules
verdict = "MALICIOUS" {
	risk_score >= 0.7
}

verdict = "SUSPICIOUS" {
	risk_score >= 0.4
	risk_score < 0.7
}
