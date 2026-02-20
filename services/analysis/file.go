package analysis

import (
	"path/filepath"
	"strings"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// AnalyzeFile performs static analysis on file metadata.
func AnalyzeFile(task domain.SAL) domain.Signal {
	signal := domain.Signal{
		EngineName: "StaticFileAnalysis",
		SignalKey:  "FILE_RISK",
		Evidence:   make(map[string]interface{}),
	}

	var riskScore float64
	var reasons []string

	filename := ""
	if task.Metadata != nil {
		if fn, ok := task.Metadata["filename"].(string); ok {
			filename = fn
		}
	}

	// 1. Double Extension Detection (e.g., .pdf.exe)
	if filename != "" {
		ext := filepath.Ext(filename)
		if ext != "" {
			remainder := strings.TrimSuffix(filename, ext)
			secondExt := filepath.Ext(remainder)
			if secondExt != "" {
				// common dangerous second extensions
				dangerous := []string{".pdf", ".zip", ".jpg", ".txt", ".doc"}
				for _, d := range dangerous {
					if strings.ToLower(secondExt) == d {
						riskScore += 0.7
						reasons = append(reasons, "Dangerous Double Extension detected: "+secondExt+ext)
						break
					}
				}
			}
		}
	}

	// 2. Executable in Archive
	if task.Metadata != nil {
		if parentType, ok := task.Metadata["parent_archive_type"].(string); ok && parentType == "zip" {
			ext := filepath.Ext(filename)
			execExts := []string{".exe", ".scr", ".vbs", ".js", ".bat", ".cmd", ".ps1"}
			for _, e := range execExts {
				if strings.ToLower(ext) == e {
					riskScore += 0.6
					reasons = append(reasons, "Executable file inside Zip archive")
					break
				}
			}
		}
	}

	// 3. Hash Reputation (Placeholder for VirusTotal)
	if sha256, ok := task.Metadata["sha256"].(string); ok {
		signal.Evidence["sha256"] = sha256
		// Placeholder: In a real system, query VT or local malware DB here.
	}

	// Cap risk score at 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	signal.Confidence = 0.95
	signal.Weight = riskScore
	signal.Evidence["reasons"] = reasons

	return signal
}
