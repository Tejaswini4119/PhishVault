$ErrorActionPreference = "Stop"

$BaseUrl = "http://localhost:8080"
$Username = "testuser_$(Get-Random)"
$Password = "testpassword123"

Write-Host "1. Registering User: $Username"
$RegisterBody = @{
    username = $Username
    password = $Password
} | ConvertTo-Json

try {
    $RegisterResp = Invoke-RestMethod -Uri "$BaseUrl/auth/register" -Method Post -Body $RegisterBody -ContentType "application/json"
    Write-Host "Success: Registered" -ForegroundColor Green
} catch {
    Write-Error "Failed to register: $_"
}

Write-Host "`n2. Logging In"
try {
    $LoginResp = Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $RegisterBody -ContentType "application/json"
    $Token = $LoginResp.token
    if ($Token) {
        Write-Host "Success: Logged in, Token received" -ForegroundColor Green
    } else {
        Write-Error "Login failed: No token"
    }
} catch {
    Write-Error "Failed to login: $_"
}

Write-Host "`n3. Accessing Protected Endpoint (Scans) WITHOUT Token"
try {
    Invoke-RestMethod -Uri "$BaseUrl/scans" -Method Get | Out-Null
    Write-Error "FAIL: Endpoint should be protected (401 expected)"
} catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 401) {
        Write-Host "Success: 401 Unauthorized received" -ForegroundColor Green
    } else {
        Write-Error "FAIL: Unexpected status code: $($_.Exception.Response.StatusCode.value__)"
    }
}

Write-Host "`n4. Accessing Protected Endpoint (Scans) WITH Token"
try {
    $Headers = @{
        Authorization = "Bearer $Token"
    }
    Invoke-RestMethod -Uri "$BaseUrl/scans" -Method Get -Headers $Headers | Out-Null
    Write-Host "Success: Access granted with token" -ForegroundColor Green
} catch {
    Write-Error "Failed to access protected resource: $_"
}
