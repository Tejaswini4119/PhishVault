package phishvault.policy

default verdict = "SAFE"
default risk_score = 0.0

# Calculate Risk Score
risk_score = score {
	score := (visual_score * 0.4) + (urgency_score * 0.2) + (intent_score * 0.3) + base_risk
}

# Factors
visual_score = input.visual_match_score
urgency_score = input.urgency_score

intent_score = 1.0 { input.intent == "CredentialHarvesting" } 
else = 0.8 { input.intent == "MalwareDistribuition" }
else = 0.0

# Boost risk if visual match + login form exists
base_risk = 0.2 { input.has_login_form == true; input.visual_match_score > 0.5 } else = 0.0

# Verdict Rules
verdict = "MALICIOUS" {
	risk_score >= 0.7
}

verdict = "SUSPICIOUS" {
	risk_score >= 0.4
	risk_score < 0.7
}
