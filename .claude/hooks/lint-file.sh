#!/bin/bash

# Extract file path from stdin JSON
FILE=$(cat | jq -r ".tool_input.file_path")

# Change to project root before running make
cd "$(git rev-parse --show-toplevel)" || exit 1

# Run make lint-file and capture output
LINT_OUTPUT=$(make lint-file FILE="$FILE" 2>&1)

# If there's output, return it as additional context
if [ -n "$LINT_OUTPUT" ]; then
  jq -n --arg ctx "$LINT_OUTPUT" '{
    hookSpecificOutput: {
      hookEventName: "PostToolUse",
      additionalContext: $ctx
    }
  }'
else
  echo "{}"
fi
