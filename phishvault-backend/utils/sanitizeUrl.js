// utils/sanitizeUrl.js

export function sanitizeUrl(url) {
  const cleaned = new URL(url.trim());
  return cleaned.href;
}
// This function takes a URL string, trims any whitespace, and creates a new URL object.
// It returns the sanitized URL as a string.
// If the URL is invalid, it will throw an error, which should be handled by the caller.
// This ensures that the URL is well-formed and can be safely used in the scanning process.
// The function can be used in the scanController to sanitize user-provided URLs before processing them.
// It helps prevent issues with malformed URLs and ensures consistent URL formatting.
// Example usage:
// const sanitizedUrl = sanitizeUrl('  http://example.com/path?query=123  ');
// console.log(sanitizedUrl); // Outputs: 'http://example.com/path?query=123'
// This function is essential for maintaining the integrity of the scanning process and ensuring that URLs are processed correctly.
// It can be extended in the future to include additional sanitization logic if needed, such as
// checking for allowed protocols or domains, or removing unwanted query parameters.