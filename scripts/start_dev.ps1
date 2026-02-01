# PhishVault-2.0 Dev Start Script (Windows)
# Usage: .\scripts\start_dev.ps1

Write-Host ">>> Starting PhishVault 2.0 Development Environment..." -ForegroundColor Cyan

# 0. Check Docker
Write-Host "   Checking Docker status..."
docker info > $null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker is NOT running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# 1. Start Infrastructure
Write-Host "`n[1/3] Starting Infrastructure (Docker)..." -ForegroundColor Yellow
Set-Location -Path "$PSScriptRoot/../deploy"
docker-compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to start Docker services." -ForegroundColor Red
    exit 1
}
Write-Host "   Infrastructure (Postgres, RabbitMQ, MinIO, Neo4j) is up." -ForegroundColor Green

# Wait for ports
Start-Sleep -Seconds 5

# 2. Start Services
Write-Host "`n[2/4] Starting Services..." -ForegroundColor Yellow
$projectRoot = "$PSScriptRoot/../"

# Ingestion Service
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$projectRoot'; Write-Host '>>> INGESTION SERVICE'; go run ./services/ingestion" -WorkingDirectory $projectRoot

# Worker Service
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$projectRoot'; Write-Host '>>> WORKER SERVICE'; go run ./services/worker" -WorkingDirectory $projectRoot

# 3. Start UI
Write-Host "`n[3/4] Starting Frontend..." -ForegroundColor Yellow
$uiPath = "$PSScriptRoot/../ui"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$uiPath'; Write-Host '>>> FRONTEND UI'; npm run dev" -WorkingDirectory $uiPath

Write-Host "`n>>> Development Environment is Ready!" -ForegroundColor Cyan
Write-Host "   - Ingestion API: http://localhost:8080"
Write-Host "   - Frontend UI:   http://localhost:3000"
Write-Host "   - Neo4j Panel:   http://localhost:7474"
