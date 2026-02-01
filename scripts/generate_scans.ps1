$urls = @(
    "http://suspicious-bank-login.com",
    "https://google-security-verify.net",
    "http://paypal-resolution-center.org",
    "https://microsoft-office365-update.xyz",
    "http://amazon-prime-delivery-status.info",
    "https://netflix-payment-declined.biz",
    "http://facebook-security-check.com",
    "https://apple-id-locked.net",
    "http://wells-fargo-verify.org",
    "https://chase-bank-alert.com"
)

Write-Host "Submitting 10 scans to PhishVault..."

foreach ($url in $urls) {
    try {
        $body = @{ url = $url } | ConvertTo-Json
        $response = Invoke-RestMethod -Uri "http://localhost:8080/submit" -Method Post -Body $body -ContentType "application/json"
        Write-Host "Submitted: $url | ID: $($response.scan_id)"
    } catch {
        Write-Host "Failed to submit $url : $_"
    }
}

Write-Host "Done! Check the dashboard."
