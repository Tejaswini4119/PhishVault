#!/bin/bash

# ════════════════════════════════════════
# PhishVault Backend Startup Script
# Author: PardhuVarma
# Date: 19 June 2025
# Description: Starts backend and MongoDB containers, checks env & assets.
# ════════════════════════════════════════

# Project Constants
BACKEND_DIR="/home/zenrage2025/Desktop/PhishVault/phishvault-backend"
SCREENSHOT_DIR="$BACKEND_DIR/screenshots"
ENV_FILE="$BACKEND_DIR/.env"
MONGO_CONTAINER_NAME_1="mongo-phishvault"
MONGO_CONTAINER_NAME_2="phishvault-backend"
MONGO_PORT=27019
BACKEND_PORT=4002
COMPOSE_FILE="$BACKEND_DIR/docker-compose.yml"
PROJECT_NAME="phishvault-backend"
BUILD_FLAG=$1

echo "🔧 [1/4] Validating backend environment..."

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
  echo "❌ .env file not found in $BACKEND_DIR"
  exit 1
fi

# Check and create screenshots directory if not exists
if [ ! -d "$SCREENSHOT_DIR" ]; then
  echo "📁 Creating screenshots directory..."
  mkdir -p "$SCREENSHOT_DIR"
fi

echo "🛠️ [2/4] Preparing backend environment..."
cd "$BACKEND_DIR" || {
  echo "❌ Failed to change directory to $BACKEND_DIR"
  exit 1
}

# Step 3: Start docker containers using docker-compose
echo "🐳 [3/4] Starting MongoDB + Backend containers..."
# Fast startup using cached image
docker-compose -p phishvault-backend up -d
if [ $? -ne 0 ]; then
  echo "❌ Failed to start Docker containers. Please check your Docker setup."
  exit 1
fi

if [ "$BUILD_FLAG" == "--build" ]; then
  echo "🔧 Building containers before startup..."
  docker-compose -p $PROJECT_NAME -f $COMPOSE_FILE up -d --build
fi

# Step 4: Run backend server
echo "🚀 [4/4] Starting PhishVault backend on port $BACKEND_PORT..."
echo "🎉 PhishVault backend is up and running! 🚀"
node server.js
if [ $? -ne 0 ]; then
  echo "❌ Failed to start the backend server. Please check your Node.js setup."
  exit 1
fi
# thank you for using PhishVault!
# End of script. 