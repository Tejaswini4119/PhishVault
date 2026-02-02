# Dev Script for PhishVault 2.0

Write-Host ">>> Starting PhishVault 2.0 Development Environment..." -ForegroundColor Cyan

# 0. Pre-flight Check
Write-Host "   Checking Docker status..."
docker info > $null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker is NOT running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}

# 1. Start Infrastructure (Docker)
Write-Host "`n[1/3] Starting Infrastructure (Docker)..." -ForegroundColor Yellow
Set-Location -Path "$PSScriptRoot/../deploy"
docker-compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to start Docker services." -ForegroundColor Red
    exit 1
}
Write-Host "   Infrastructure (Postgres, RabbitMQ, MinIO, Neo4j) is up." -ForegroundColor Green

# Wait for Neo4j HTTP to be ready (more reliable than just port open)
Write-Host "   Waiting for Neo4j HTTP (7474) and RabbitMQ (5672)..." -ForegroundColor Yellow
$timeout = 60
$sw = [System.Diagnostics.Stopwatch]::StartNew()

while ($sw.Elapsed.TotalSeconds -lt $timeout) {
    # Check RabbitMQ Port
    $rabbitmq = Test-NetConnection -ComputerName localhost -Port 5672 -InformationLevel Quiet
    
    # Check Neo4j HTTP (Status 200)
    $neo4jReady = $false
    try {
        $resp = Invoke-WebRequest -Uri "http://localhost:7474" -Method Head -ErrorAction SilentlyContinue
        if ($resp.StatusCode -eq 200) { $neo4jReady = $true }
    } catch {}

    if ($rabbitmq -and $neo4jReady) {
        Write-Host "`n   Ports are open and HTTP is responding!" -ForegroundColor Green
        Write-Host "   Waiting 10s for internal initialization..." -ForegroundColor Cyan
        Start-Sleep -Seconds 10
        break
    }
    Write-Host "." -NoNewline -ForegroundColor Yellow
    Start-Sleep -Seconds 2
}
if ($sw.Elapsed.TotalSeconds -ge $timeout) {
    Write-Host "`n   ⚠️  Timed out waiting for services. Proceeding explicitly..." -ForegroundColor Yellow
}

# 2. Start Backend (Ingestion API)
Write-Host "`n[2/4] Starting Backend Services..." -ForegroundColor Yellow
$backendPath = "$PSScriptRoot/../"

# Start Ingestion API
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$backendPath'; Write-Host 'Starting Ingestion API...'; go run ./services/ingestion" -WorkingDirectory $backendPath

# Start Analysis Worker
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$backendPath'; Write-Host 'Starting Analysis Worker...'; go run ./services/worker/main.go" -WorkingDirectory $backendPath

Write-Host "   Backend & Worker started in new windows." -ForegroundColor Green

# 3. Start Frontend (Next.js UI)
Write-Host "`n[3/4] Starting Frontend UI..." -ForegroundColor Yellow
$uiPath = "$PSScriptRoot/../ui"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$uiPath'; Write-Host 'Starting Next.js UI...'; npm.cmd run dev" -WorkingDirectory $uiPath
Write-Host "   Frontend started in new window." -ForegroundColor Green

Write-Host "`n>>> Development Environment is Ready!" -ForegroundColor Cyan
Write-Host "   - Backend: http://localhost:8080"
Write-Host "   - Frontend: http://localhost:3000"
Write-Host "   - MinIO: http://localhost:9001"
Write-Host "   - Neo4j: http://localhost:7474"
