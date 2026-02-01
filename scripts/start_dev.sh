#!/bin/bash

# PhishVault-2.0 Dev Start Script
# Usage: ./scripts/start_dev.sh

echo ">>> Starting PhishVault 2.0 Development Environment..."

# Resolve Project Root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT" || { echo "Failed to cd to project root"; exit 1; }
echo ">>> Working Directory: $(pwd)"

# Function to run command in new tmux pane
run_in_tmux() {
    local cmd=$1
    tmux split-window -h "$cmd"
    tmux select-layout tiled
}

# 1. Start Infrastructure
echo "[1/4] Starting Infrastructure..."
cd deploy || exit
docker compose up -d
cd ..

if [ -n "$TMUX" ]; then
    echo ">>> Tmux detected. Configuring panes..."
    
    # We are in a tmux pane. Let's split it up.
    # Pane 1 (Current): Will show this script status and then maybe become one of the logs
    
    # Split for Ingestion (Pane 2)
    tmux split-window -h "cd services/ingestion && echo 'Starting Ingestion...' && go run .; bash"
    
    # Split for Worker (Pane 3)
    tmux split-window -v "cd services/worker && echo 'Starting Worker...' && go run .; bash"
    
    # Split for UI (Pane 4)
    tmux select-pane -t 0
    tmux split-window -v "cd ui && echo 'Starting UI...' && export NVM_DIR=\"\$HOME/.nvm\" && . \"\$NVM_DIR/nvm.sh\" && npm run dev; bash"
    
    # Arrange
    tmux select-layout tiled
    
    echo ">>> Services starting in panes."
    echo "    - Top Left: Docker Status / Control"
    echo "    - Top Right: Ingestion API (:8080)"
    echo "    - Bot Right: Worker Service"
    echo "    - Bot Left: Web UI (:3000)"

else
    echo ">>> Not in Tmux. Starting in background..."
    
    (cd services/ingestion && go run .) &
    PID_INGEST=$!
    
    (cd services/worker && go run .) &
    PID_WORKER=$!
    
    (cd ui && export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh" && npm run dev) &
    PID_UI=$!
    
    echo "Services started with PIDs: Ingestion($PID_INGEST), Worker($PID_WORKER), UI($PID_UI)"
    echo "Press Ctrl+C to stop all."
    
    trap "kill $PID_INGEST $PID_WORKER $PID_UI; exit" SIGINT
    wait
fi
