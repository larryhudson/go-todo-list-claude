#!/bin/bash

# SessionStart hook to set up development environment
# This runs when Claude Code starts a new session

cd "$CLAUDE_PROJECT_DIR" || exit 1

# Set Go environment variables for the session
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export GOTOOLCHAIN=local' >> "$CLAUDE_ENV_FILE"
  echo "✓ Go environment configured (GOTOOLCHAIN=local)"
fi

# Install hivemind if not already installed
if ! command -v hivemind &>/dev/null; then
  echo "Installing hivemind..."
  go install github.com/DarthSim/hivemind@latest &>/dev/null
  echo "✓ hivemind installed successfully"
else
  echo "✓ hivemind already installed"
fi

# Check if we're in a project with a frontend directory
if [ -d "frontend" ]; then
  cd frontend || exit 1

  # Check if tsgo is already installed
  if ! npm list @typescript/native-preview &>/dev/null; then
    echo "Installing @typescript/native-preview (tsgo)..."
    npm install --save-dev @typescript/native-preview &>/dev/null
    echo "✓ tsgo installed successfully"
  else
    echo "✓ tsgo already installed"
  fi
fi

# Return success
echo "{}"
