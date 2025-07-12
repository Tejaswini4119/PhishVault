#!/bin/bash

# ════════════════════════════════════════════════
# PhishVault Frontend Startup Script
# Author: PardhuVarma
# Date: 12 July 2025
# Description: Starts frontend development server (React)
# ════════════════════════════════════════════════

# Project Constants
FRONTEND_DIR="/home/zenrage2025/Desktop/PhishVault/phishvault-frontend"
PORT=3000

# Step 1: Navigate to frontend directory
echo "📁 Navigating to frontend directory..."
cd "$FRONTEND_DIR" || {
  echo "❌ Failed to access frontend directory: $FRONTEND_DIR"
  exit 1
}

# Step 2: Check for Node.js
if ! command -v node &> /dev/null; then
  echo "❌ Node.js is not installed. Please install it to run the frontend."
  exit 1
fi

# Step 3: Check for node_modules (if not installed, install them)
if [ ! -d "node_modules" ]; then
  echo "📦 Installing frontend dependencies..."
  npm install
fi

# Step 4: Start development server
echo "🚀 Starting React development server on http://localhost:$PORT..."
npm start
