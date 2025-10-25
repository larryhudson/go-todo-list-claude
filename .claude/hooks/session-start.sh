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

# Start the dev server automatically
echo "Starting dev server..."
if make dev 2>&1 | grep -q "already running"; then
  echo "✓ Dev server already running"
else
  echo "✓ Dev server started"
  echo "  Use 'make dev-logs' to view logs"
  echo "  Use 'make dev-stop' to stop the server"
fi

# Return success
echo "{}"
