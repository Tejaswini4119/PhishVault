// services/threatScorer.js

export default function threatScorer({ logs, redirects, cookies }) {
  let score = 0;
  let notes = [];

  if (redirects.length > 3) {
    score += 2;
    notes.push('Multiple redirects');
  }

  const suspiciousPatterns = ['eval', 'atob', 'obfuscate', 'decodeURIComponent'];
  const suspiciousLogs = logs.filter(log => suspiciousPatterns.some(p => log.includes(p)));

  if (suspiciousLogs.length) {
    score += 3;
    notes.push('Suspicious JS logs');
  }

  if (cookies.length > 5) {
    score += 1;
    notes.push('Excessive cookies');
  }

  let verdict = 'Safe';
  if (score >= 6) verdict = 'Malicious';
  else if (score >= 3) verdict = 'Suspicious';

  return { score, verdict, notes };
}
// This function analyzes the scan data and assigns a score based on various heuristics.
// It checks for multiple redirects, suspicious JavaScript logs, and excessive cookies.
// The score is used to determine the verdict: 'Safe', 'Suspicious', or 'Malicious'.
// The function returns an object containing the score, verdict, and any notes on the