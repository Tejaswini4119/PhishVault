export default function threatScorer({ logs, redirects, cookies, html, url }) {
  let score = 0;
  let notes = [];

  // 1. Redirects
  if (redirects.length > 3) {
    score += 2;
    notes.push('Multiple redirects');
  }

  // 2. Suspicious JS logs
  const suspiciousPatterns = ['eval', 'atob', 'obfuscate', 'decodeURIComponent'];
  const suspiciousLogs = logs.filter(log =>
    suspiciousPatterns.some(p => log.includes(p))
  );
  if (suspiciousLogs.length) {
    score += 3;
    notes.push('Suspicious JavaScript behavior detected in logs');
  }

  // 3. Excessive cookies
  if (cookies.length > 5) {
    score += 1;
    notes.push('Excessive number of cookies set');
  }

  // 4. Password field check
  if (/type\s*=\s*["']?password["']?/i.test(html)) {
    score += 3;
    notes.push('Password field found');
  }

  // 5. Form action domain mismatch
  const formActionMatch = html.match(/<form[^>]+action=["']?([^"'>]+)["']?/i);
  if (formActionMatch && formActionMatch[1]) {
    try {
      const baseDomain = new URL(url).hostname;
      const formDomain = new URL(formActionMatch[1], url).hostname;
      if (baseDomain !== formDomain) {
        score += 3;
        notes.push(`Form action points to external domain: ${formDomain}`);
      }
    } catch {
      // malformed form URL, ignore
    }
  }

  // 6. Phishing keywords in HTML
  const lowered = html.toLowerCase();
  const keywords = ['login', 'sign in', 'verify', 'account', 'reset password', 'update info', 'confirm'];
  keywords.forEach(word => {
    const pattern = new RegExp(`\\b${word}\\b`, 'i');
    if (pattern.test(lowered)) {
      score += 1;
      notes.push(`Suspicious keyword found: "${word}"`);
    }
  });

  // Final Verdict
  let verdict = 'Safe';
  if (score >= 7) verdict = 'Malicious';
  else if (score >= 4) verdict = 'Suspicious';

  return { score, verdict, notes };
}
// threatScorer.js