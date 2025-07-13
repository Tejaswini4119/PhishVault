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

  // 4. Password and Credential fields
  if (/type\s*=\s*["']?password["']?/i.test(html)) {
    score += 3;
    notes.push('Password field found');
  }
  if (/name\s*=\s*["']?(username|email|user)["']?/i.test(html)) {
    score += 2;
    notes.push('Credential input field found');
  }
  if (/type=["']hidden["'].*name=["']?(token|auth|csrf)/i.test(html)) {
    score += 2;
    notes.push('Hidden auth-related fields detected');
  }

  // 5. Form action domain mismatch or insecure submission
  const formActionMatch = html.match(/<form[^>]+action=["']?([^"'>]+)["']?/i);
  if (formActionMatch && formActionMatch[1]) {
    try {
      const baseDomain = new URL(url).hostname;
      const formDomain = new URL(formActionMatch[1], url).hostname;
      if (baseDomain !== formDomain) {
        score += 3;
        notes.push(`Form action points to external domain: ${formDomain}`);
      }
      if (formActionMatch[1].startsWith('http://')) {
        score += 3;
        notes.push('Form submission is not secure (HTTP)');
      }
    } catch {
      // malformed form URL, ignore
    }
  }

  // 6. Phishing keywords
  const lowered = html.toLowerCase();
  const keywords = ['login', 'sign in', 'verify', 'account', 'reset password', 'update info', 'confirm'];
  keywords.forEach(word => {
    const pattern = new RegExp(`\\b${word}\\b`, 'i');
    if (pattern.test(lowered)) {
      score += 1;
      notes.push(`Suspicious keyword found: "${word}"`);
    }
  });

  // 7. Cloaking / Fingerprinting detection
  const cloakingKeywords = [
    'navigator.userAgent',
    'navigator.plugins',
    'screen.width',
    'navigator.webdriver',
    'Intl.DateTimeFormat',
    'timezoneOffset'
  ];
  if (cloakingKeywords.some(k => html.includes(k))) {
    score += 2;
    notes.push('Potential fingerprinting or bot detection scripts found');
  }

  // 8. Suspicious external JS
  const badScriptMatch = html.match(/<script[^>]+src=["']([^"']+)["']/gi) || [];
  badScriptMatch.forEach(scriptTag => {
    if (/(\.xyz|\.tk|\.ru|\.pw|\.click|dropbox|pastebin|googledrive|http:\/\/\d+\.\d+\.\d+\.\d+)/i.test(scriptTag)) {
      score += 2;
      notes.push('Suspicious external JS source found');
    }
  });

  // 9. Brand impersonation
  const brands = ['netflix', 'paypal', 'microsoft', 'amazon', 'bank', 'apple', 'google'];
  brands.forEach(brand => {
    if (html.toLowerCase().includes(brand)) {
      score += 1;
      notes.push(`Brand impersonation indicator: "${brand}"`);
    }
  });

  // 10. Anti-analysis / delay logic
  if (/debugger\s*;/.test(html) || /while\s*\(\s*true\s*\)/.test(html)) {
    score += 3;
    notes.push('Anti-analysis or infinite loop behavior detected');
  }
  if (/setTimeout\s*\([^,]+,\s*(1\d{4,}|[2-9]\d{4,})\)/.test(html)) {
    score += 2;
    notes.push('Suspicious delayed execution script detected');
  }

  // Final Verdict
  let verdict = 'Safe';
  if (score >= 7) verdict = 'Malicious';
  else if (score >= 4) verdict = 'Suspicious';

  return {
    score,
    verdict,
    notes,
    details: notes.join('; ')
  };
}
