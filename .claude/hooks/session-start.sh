#!/bin/bash

# SessionStart hook to set up development environment
# This runs when Claude Code starts a new session

cd "$CLAUDE_PROJECT_DIR" || exit 1

# Set Go environment variables for the session
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export GOTOOLCHAIN=local' >> "$CLAUDE_ENV_FILE"
  # Add Go bin directory to PATH for go-installed tools
  GOBIN=$(go env GOPATH)/bin
  echo "export PATH=\"\$PATH:$GOBIN\"" >> "$CLAUDE_ENV_FILE"
  echo "✓ Go environment configured (GOTOOLCHAIN=local, PATH includes Go bin)"
fi

# Install all dependencies (backend and frontend)
echo "Installing dependencies..."
if make install &>/dev/null; then
  echo "✓ All dependencies installed successfully"
else
  echo "⚠ Warning: Some dependencies may not have installed correctly"
fi

# Return success
echo "{}"
