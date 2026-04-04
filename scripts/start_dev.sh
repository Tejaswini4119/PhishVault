#!/bin/bash

# PhishVault-2.0 Dev Start Script
# Usage: ./scripts/start_dev.sh

set -u

echo ">>> Starting PhishVault 2.0 Development Environment..."

# Resolve Project Root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT" || { echo "Failed to cd to project root"; exit 1; }
echo ">>> Working Directory: $(pwd)"

COMPOSE_FILE="$PROJECT_ROOT/deploy/docker-compose.yml"

compose() {
    docker compose -f "$COMPOSE_FILE" "$@"
}

wait_for_service() {
    local service="$1"
    local timeout_secs="$2"
    local waited=0

    local cid
    cid="$(compose ps -q "$service" 2>/dev/null)"
    if [ -z "$cid" ]; then
        echo ">>> ERROR: service '$service' has no container id"
        return 1
    fi

    while [ "$waited" -lt "$timeout_secs" ]; do
        local status
        status="$(docker inspect -f '{{.State.Status}}' "$cid" 2>/dev/null || echo "unknown")"

        if [ "$status" = "running" ]; then
            local health
            health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$cid" 2>/dev/null || echo "none")"
            if [ "$health" = "healthy" ] || [ "$health" = "none" ]; then
                echo ">>> $service is ready ($status/$health)"
                return 0
            fi
        elif [ "$status" = "exited" ] || [ "$status" = "dead" ]; then
            echo ">>> ERROR: $service is $status"
            compose logs "$service" --tail=60
            return 1
        fi

        sleep 2
        waited=$((waited + 2))
    done

    echo ">>> ERROR: timed out waiting for $service (${timeout_secs}s)"
    compose ps "$service"
    compose logs "$service" --tail=40
    return 1
}

apply_db_schema() {
    echo ">>> Applying DB schema..."
    if ! compose exec -T postgres psql -U phishvault -d phishvault -v ON_ERROR_STOP=1 < "$PROJECT_ROOT/deploy/schema.sql"; then
        echo ">>> WARNING: failed to apply schema from deploy/schema.sql"
    fi
}

check_playwright() {
    local browser_cache="${PLAYWRIGHT_BROWSERS_PATH:-$HOME/.cache/ms-playwright}"
    if [ ! -d "$browser_cache" ] || ! find "$browser_cache" -maxdepth 1 -type d -name 'chromium-*' | grep -q .; then
        echo ">>> WARNING: Playwright browser driver is not installed."
        echo ">>> Run once to fix worker browser scans:"
        echo "    cd $PROJECT_ROOT && go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5200.1 install --with-deps chromium"
    fi
}

build_ui_command() {
    cat <<'EOF'
export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then
  . "$NVM_DIR/nvm.sh"
elif ! command -v node >/dev/null 2>&1; then
  echo "UI startup failed: neither nvm nor system node found"
  exit 1
fi
if ! command -v npm >/dev/null 2>&1; then
    echo "UI startup failed: npm is not available in PATH"
    exit 1
fi
if [ ! -x "node_modules/.bin/next" ]; then
    echo "Installing UI dependencies..."
    npm install --no-audit --no-fund
fi
npm run dev
EOF
}

# 1. Start Infrastructure
echo "[1/4] Starting Infrastructure..."
compose up -d --remove-orphans

echo ">>> Waiting for infrastructure readiness..."
wait_for_service postgres 60 || exit 1
wait_for_service rabbitmq 60 || exit 1
wait_for_service minio 60 || exit 1

# Neo4j is non-blocking for local startup but we report detailed diagnostics.
if ! wait_for_service neo4j 90; then
    echo ">>> WARNING: Neo4j failed to become ready. Ingestion/analysis will run with degraded graph mode."
    echo ">>> Tip: if this persists, reset only Neo4j state with: docker compose -f deploy/docker-compose.yml down && docker volume rm deploy_neo4j_data"
fi

apply_db_schema
check_playwright

UI_CMD="$(build_ui_command)"

if [ -n "${TMUX:-}" ]; then
    echo ">>> Tmux detected. Configuring panes..."

    tmux split-window -h "cd '$PROJECT_ROOT/services/ingestion' && echo 'Starting Ingestion...' && go run .; bash"
    tmux split-window -v "cd '$PROJECT_ROOT/services/worker' && echo 'Starting Worker...' && go run .; bash"
    tmux select-pane -t 0
    tmux split-window -v "cd '$PROJECT_ROOT/ui' && echo 'Starting UI...' && $UI_CMD; bash"

    tmux select-layout tiled

    echo ">>> Services starting in panes."
    echo "    - Top Left: Docker Status / Control"
    echo "    - Top Right: Ingestion API (:8080)"
    echo "    - Bot Right: Worker Service"
    echo "    - Bot Left: Web UI (:3000)"
else
    echo ">>> Not in Tmux. Starting in background..."

    (cd "$PROJECT_ROOT/services/ingestion" && go run .) &
    PID_INGEST=$!

    (cd "$PROJECT_ROOT/services/worker" && go run .) &
    PID_WORKER=$!

    (cd "$PROJECT_ROOT/ui" && eval "$UI_CMD") &
    PID_UI=$!

    echo "Services started with PIDs: Ingestion($PID_INGEST), Worker($PID_WORKER), UI($PID_UI)"
    echo "Press Ctrl+C to stop all."

    trap 'kill "$PID_INGEST" "$PID_WORKER" "$PID_UI" 2>/dev/null; exit' SIGINT SIGTERM
    wait
fi
