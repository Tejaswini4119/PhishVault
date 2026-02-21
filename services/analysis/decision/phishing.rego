package phishvault.policy

default verdict = "SAFE"
default risk_score = 0.0

# --- Multipliers ---

domain_multiplier = 1.6 { input.domain_age_days < 14 }
else = 1.3 { input.domain_age_days < 30 }
else = 1.0

brand_multiplier = 1.5 { input.visual_match_score > 0.8; input.has_login_form }
else = 1.1 { input.visual_match_score > 0.5 }
else = 1.0

form_multiplier = 1.2 { input.has_login_form; input.intent == "CredentialHarvesting" }
else = 1.0

# --- Core Logic ---

base_score = score {
	score := (input.visual_match_score * 0.3) + (input.urgency_score * 0.2) + (intent_score * 0.5)
}

intent_score = 0.8 { input.intent == "CredentialHarvesting" } 
else = 0.6 { input.intent == "MalwareDistribuition" }
else = 0.0

raw_risk = score {
	score := base_score * domain_multiplier * brand_multiplier * form_multiplier
}

# Final Risk Score (0.0 to 1.0)
risk_score = 1.0 { raw_risk > 1.0 }
else = raw_risk

# --- Verdict Rules ---
MALICIOUS_THRESHOLD = 0.7
SUSPICIOUS_THRESHOLD = 0.45

verdict = "MALICIOUS" {
	risk_score >= MALICIOUS_THRESHOLD
}

verdict = "SUSPICIOUS" {
	risk_score >= SUSPICIOUS_THRESHOLD
	risk_score < MALICIOUS_THRESHOLD
}
