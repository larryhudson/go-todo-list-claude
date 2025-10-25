# Claude Code Development Process Guide

A language-agnostic template for setting up an optimal development workflow with Claude Code.

## Overview

This guide demonstrates techniques to create a **tight feedback loop** between Claude Code and your project. The goal is to enable the AI assistant to quickly detect and fix its own mistakes through:

1. **Unified logging** - See all output in one place
2. **Automatic linting** - Catch errors immediately after file changes
3. **Auto-starting dev server** - Begin development instantly
4. **Clear guidelines** - Provide project-specific instructions

## Core Components

### 1. Process Aggregation with Unified Logging

**Problem**: Multi-service projects (backend + frontend, microservices) produce logs in separate terminals, making it hard to observe the system's state.

**Solution**: Use a process manager to run all services concurrently and aggregate their logs into a single file.

#### Recommended Tools

- **Hivemind** (Ruby) - Simple, no config needed beyond Procfile
- **Foreman** (Ruby) - Original Procfile runner
- **overmind** (Go) - Hivemind alternative with tmux integration
- **honcho** (Python) - Python port of Foreman
- **docker-compose** - If you're already using containers

#### Setup Pattern

**Step 1**: Create a `Procfile` listing your processes:

```procfile
# Procfile - defines processes to run concurrently
backend: [command to run backend]
frontend: [command to run frontend]
worker: [command to run background worker]
```

**Language-Specific Examples**:

```procfile
# Go + React
backend: go run ./cmd/server/main.go
frontend: cd frontend && npm run dev

# Python + Vue
backend: python manage.py runserver
frontend: cd frontend && npm run serve

# Node.js + Next.js
api: cd api && npm run dev
web: cd web && npm run dev

# Ruby on Rails + React
rails: bundle exec rails server
webpack: bin/webpack-dev-server

# Java + Angular
backend: ./gradlew bootRun
frontend: cd frontend && ng serve
```

**Step 2**: Create Makefile targets for development commands:

```makefile
# Makefile

.PHONY: dev dev-logs dev-stop

dev: ## Start all services with unified logging
	@if [ -f .dev-server.pid ]; then \
		PID=$$(cat .dev-server.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Dev server is already running (PID: $$PID)"; \
			exit 1; \
		fi; \
	fi
	@mkdir -p .logs
	@nohup hivemind Procfile > dev-server.log 2>&1 & echo $$! > .dev-server.pid
	@echo "Dev server started (PID: $$(cat .dev-server.pid))"
	@echo "Run 'make dev-logs' to view logs"

dev-logs: ## Tail the unified dev server logs
	@if [ ! -f dev-server.log ]; then \
		echo "No dev server log file found. Has the dev server been started?"; \
		exit 1; \
	fi
	tail -f dev-server.log | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g'

dev-stop: ## Stop the dev server
	@if [ ! -f .dev-server.pid ]; then \
		echo "No dev server PID file found"; \
		exit 1; \
	fi
	@PID=$$(cat .dev-server.pid); \
	if ps -p $$PID > /dev/null 2>&1; then \
		echo "Stopping dev server (PID: $$PID)..."; \
		kill $$PID; \
		rm .dev-server.pid; \
		echo "Dev server stopped"; \
	else \
		echo "Dev server process not found, cleaning up stale PID file"; \
		rm .dev-server.pid; \
	fi
```

**Step 3**: Add to `.gitignore`:

```gitignore
# Development server
.dev-server.pid
dev-server.log
.logs/
```

**Key Benefits**:
- Single command starts everything: `make dev`
- All logs in one file: `make dev-logs`
- PID tracking prevents duplicate servers
- Works with any language/framework

---

### 2. Automatic Linting with Hooks

**Problem**: Claude Code may introduce syntax errors, style violations, or type errors. Without immediate feedback, these accumulate and require manual detection.

**Solution**: Configure hooks that automatically lint/format files after Claude edits them.

#### Setup Pattern

**Step 1**: Create `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "./.claude/hooks/session-start.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "./.claude/hooks/lint-file.sh"
          }
        ]
      }
    ]
  }
}
```

**Step 2**: Create `.claude/hooks/session-start.sh`:

```bash
#!/bin/bash
# Runs once when a Claude Code session starts

cd "$CLAUDE_PROJECT_DIR" || exit 1

# Example: Set environment variables for this session
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export MY_ENV_VAR=value' >> "$CLAUDE_ENV_FILE"
fi

# Example: Install dependencies
echo "Installing dependencies..."
make install 2>&1 | grep -E "✓|Error" || echo "✓ Dependencies ready"

# Example: Start dev server automatically
echo "Starting dev server..."
if make dev 2>&1 | grep -q "already running"; then
  echo "✓ Dev server already running"
else
  echo "✓ Dev server started"
  echo "  Use 'make dev-logs' to view logs"
fi
```

**Step 3**: Create `.claude/hooks/lint-file.sh`:

```bash
#!/bin/bash
# Runs after Edit or Write tool is used

# Extract the file path from Claude's tool input
FILE=$(cat | jq -r ".tool_input.file_path")

# Change to project root
cd "$(git rev-parse --show-toplevel)" || exit 1

# Run linting via Makefile
LINT_OUTPUT=$(make lint-file FILE="$FILE" 2>&1)

# Return output as additional context to Claude
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
```

**Step 4**: Add language-specific linting to your Makefile:

```makefile
lint-file: ## Lint a single file (requires FILE=path/to/file)
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE variable is required"; \
		exit 1; \
	fi
	@# Detect file type and run appropriate linter
	@if echo "$(FILE)" | grep -q '\.go$$'; then \
		echo "Go linting: $(FILE)"; \
		gofmt -w -s "$(FILE)" 2>&1; \
		golangci-lint run "$(FILE)" 2>&1 || true; \
	elif echo "$(FILE)" | grep -qE '\.(ts|tsx)$$'; then \
		echo "TypeScript linting: $(FILE)"; \
		npx prettier --write "$(FILE)" 2>&1; \
		npx tsc --noEmit "$(FILE)" 2>&1 || true; \
	elif echo "$(FILE)" | grep -q '\.py$$'; then \
		echo "Python linting: $(FILE)"; \
		black "$(FILE)" 2>&1; \
		pylint "$(FILE)" 2>&1 || true; \
	elif echo "$(FILE)" | grep -qE '\.(js|jsx)$$'; then \
		echo "JavaScript linting: $(FILE)"; \
		npx prettier --write "$(FILE)" 2>&1; \
		npx eslint --fix "$(FILE)" 2>&1 || true; \
	elif echo "$(FILE)" | grep -q '\.rb$$'; then \
		echo "Ruby linting: $(FILE)"; \
		rubocop -a "$(FILE)" 2>&1 || true; \
	elif echo "$(FILE)" | grep -q '\.rs$$'; then \
		echo "Rust formatting: $(FILE)"; \
		rustfmt "$(FILE)" 2>&1; \
		cargo clippy --all-targets 2>&1 || true; \
	elif echo "$(FILE)" | grep -qE '\.(java|kt)$$'; then \
		echo "Java/Kotlin formatting: $(FILE)"; \
		# Add your formatter here (e.g., google-java-format, ktlint); \
	else \
		echo "No linter configured for $(FILE)"; \
	fi
```

**Make hooks executable**:

```bash
chmod +x .claude/hooks/*.sh
```

**Language-Specific Linter Examples**:

| Language | Formatter | Linter |
|----------|-----------|--------|
| **Go** | `gofmt`, `goimports` | `golangci-lint` |
| **Python** | `black`, `autopep8` | `pylint`, `flake8`, `mypy` |
| **JavaScript/TypeScript** | `prettier` | `eslint`, `tsc` |
| **Ruby** | `rubocop -a` | `rubocop` |
| **Rust** | `rustfmt` | `cargo clippy` |
| **Java** | `google-java-format` | `checkstyle`, `spotbugs` |
| **C/C++** | `clang-format` | `clang-tidy`, `cppcheck` |
| **C#** | `dotnet format` | `dotnet build` |
| **PHP** | `php-cs-fixer` | `phpstan`, `psalm` |

**Key Benefits**:
- Immediate feedback on syntax/style errors
- Claude sees linter output as "additional context" and can fix issues immediately
- Prevents error accumulation
- Enforces project style consistently

---

### 2.5. Layered Code Quality Strategy

**Important**: The automatic linting hooks described above are just **Layer 1** of a complete quality strategy. For production-grade code, you should implement multiple layers of checks, each with different performance characteristics and blocking behavior.

#### The Five Layers

```
┌───────────────────────────────────────────────────────────────┐
│ Layer 1: Fast Checks (Hooks)                                 │
│ ├─ Triggered: On every file change                           │
│ ├─ Speed: < 1 second                                         │
│ ├─ Blocking: NO - warnings only                              │
│ └─ Examples: Auto-formatting, basic syntax validation        │
├───────────────────────────────────────────────────────────────┤
│ Layer 2: Pre-commit Checks                                   │
│ ├─ Triggered: Before git commit                              │
│ ├─ Speed: < 30 seconds                                       │
│ ├─ Blocking: YES - commit fails if checks fail               │
│ └─ Examples: Full linting, type checking, fast unit tests    │
├───────────────────────────────────────────────────────────────┤
│ Layer 3: CI/CD Checks                                        │
│ ├─ Triggered: On git push or PR                              │
│ ├─ Speed: 1-10 minutes                                       │
│ ├─ Blocking: YES - PR cannot merge if checks fail            │
│ └─ Examples: Full test suite, security scans, builds         │
├───────────────────────────────────────────────────────────────┤
│ Layer 3.5: Automated LLM Review (Optional)                   │
│ ├─ Triggered: After CI passes                                │
│ ├─ Speed: 1-5 minutes                                        │
│ ├─ Blocking: OPTIONAL - can be informational or required     │
│ └─ Examples: Pattern consistency, simplicity, edge cases     │
├───────────────────────────────────────────────────────────────┤
│ Layer 4: Human Review                                        │
│ ├─ Triggered: During PR review                               │
│ ├─ Speed: Hours to days                                      │
│ ├─ Blocking: YES - requires approval to merge                │
│ └─ Examples: Architecture, business logic, strategic tradeoffs│
└───────────────────────────────────────────────────────────────┘
```

#### Layer 1: Fast Checks (PostToolUse Hooks)

**Purpose**: Catch obvious mistakes immediately with near-zero latency.

**Characteristics**:
- **Speed**: < 1 second (must be fast to not disrupt development)
- **Blocking**: NO (should warn but not prevent Claude from continuing)
- **Auto-fix**: YES (apply formatters automatically)

**What to include**:
- ✅ Code formatting (prettier, black, gofmt, rustfmt)
- ✅ Basic syntax validation
- ✅ Import sorting
- ✅ Trailing whitespace removal
- ✅ Simple style violations that can be auto-fixed

**What to exclude**:
- ❌ Slow type checking (> 1 second)
- ❌ Running tests
- ❌ Complex static analysis
- ❌ Compilation/build steps

**Example configuration**:

```makefile
lint-file-fast: ## Fast linting for hooks (< 1 second)
	@if echo "$(FILE)" | grep -q '\.go$$'; then \
		gofmt -w -s "$(FILE)" 2>&1; \
		# Skip golangci-lint here - too slow for hooks
	elif echo "$(FILE)" | grep -qE '\.(ts|tsx)$$'; then \
		npx prettier --write "$(FILE)" 2>&1; \
		# Skip tsc here - too slow for hooks
	elif echo "$(FILE)" | grep -q '\.py$$'; then \
		black "$(FILE)" 2>&1; \
		isort "$(FILE)" 2>&1; \
		# Skip pylint/mypy here - too slow for hooks
	fi
```

**Hook configuration** (non-blocking):

```bash
#!/bin/bash
# .claude/hooks/lint-file.sh

FILE=$(cat | jq -r ".tool_input.file_path")
cd "$(git rev-parse --show-toplevel)" || exit 1

# Run fast checks only
LINT_OUTPUT=$(make lint-file-fast FILE="$FILE" 2>&1)

# Return warnings but don't block
if [ -n "$LINT_OUTPUT" ]; then
  jq -n --arg ctx "⚠️  Quick lint feedback:\n$LINT_OUTPUT" '{
    hookSpecificOutput: {
      hookEventName: "PostToolUse",
      additionalContext: $ctx
    }
  }'
else
  echo "{}"
fi
```

---

#### Layer 2: Pre-commit Checks

**Purpose**: Prevent broken code from entering git history.

**Characteristics**:
- **Speed**: < 30 seconds (acceptable to wait before commit)
- **Blocking**: YES (commit should fail if checks fail)
- **Auto-fix**: OPTIONAL (can auto-fix and re-stage)

**What to include**:
- ✅ Full linting (eslint, pylint, golangci-lint)
- ✅ Type checking (tsc, mypy, flow)
- ✅ Unit tests (fast tests only)
- ✅ Code formatting verification
- ✅ Dependency vulnerability checks (basic)
- ✅ Secret scanning (prevent committing credentials)

**Tools**:
- **pre-commit** (Python) - Multi-language framework
- **husky** (Node.js) - Git hooks for npm projects
- **lefthook** (Go) - Fast git hooks manager
- **git hooks** - Native git hook scripts

**Example with pre-commit framework**:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: format
        name: Format code
        entry: make format-all
        language: system
        pass_filenames: false

      - id: lint
        name: Lint code
        entry: make lint-all
        language: system
        pass_filenames: false

      - id: type-check
        name: Type check
        entry: make type-check
        language: system
        pass_filenames: false

      - id: test-fast
        name: Run fast unit tests
        entry: make test-fast
        language: system
        pass_filenames: false

  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks
```

**Installation**:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Test it
pre-commit run --all-files
```

**Makefile targets for pre-commit**:

```makefile
format-all: ## Format all code
	gofmt -w -s ./...
	cd frontend && npx prettier --write .

lint-all: ## Lint all code
	golangci-lint run ./...
	cd frontend && npx eslint src/

type-check: ## Type check all code
	cd frontend && npx tsc --noEmit

test-fast: ## Run fast unit tests only
	go test -short ./...
	cd frontend && npm run test:unit
```

---

#### Layer 3: CI/CD Checks

**Purpose**: Comprehensive validation before merging to main branch.

**Characteristics**:
- **Speed**: 1-10 minutes (acceptable for CI pipeline)
- **Blocking**: YES (PR cannot merge if CI fails)
- **Coverage**: COMPREHENSIVE (test everything)

**What to include**:
- ✅ All Layer 2 checks (lint, type check, format verification)
- ✅ Full test suite (unit + integration + e2e)
- ✅ Code coverage requirements (e.g., > 80%)
- ✅ Security scanning (SAST, dependency vulnerabilities)
- ✅ Build verification (ensure it compiles/builds)
- ✅ Docker image builds
- ✅ License compliance checks
- ✅ Performance benchmarks (optional)

**Example GitHub Actions workflow**:

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run linters
        run: make lint-all

  type-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Type check
        run: make type-check

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run full test suite
        run: make test
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run security scan
        run: |
          npm audit --audit-level=high
          go list -json -m all | nancy sleuth

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build application
        run: make build
      - name: Build Docker image
        run: docker build -t myapp:${{ github.sha }} .
```

**Example GitLab CI**:

```yaml
# .gitlab-ci.yml
stages:
  - lint
  - test
  - security
  - build

lint:
  stage: lint
  script:
    - make lint-all

test:
  stage: test
  script:
    - make test
  coverage: '/coverage: \d+\.\d+/'

security:
  stage: security
  script:
    - make security-scan

build:
  stage: build
  script:
    - make build
```

---

#### Layer 3.5: Automated LLM Review

**Purpose**: Catch issues that deterministic tools miss but don't require human judgment.

**Characteristics**:
- **Speed**: 1-5 minutes (runs after CI passes)
- **Blocking**: OPTIONAL (can be informational or blocking)
- **Coverage**: CONTEXTUAL (understands codebase patterns)

**What LLM review catches**:
- ✅ **Unnecessary complexity** - Is this the simplest approach?
- ✅ **Pattern violations** - Does this follow existing patterns in the codebase?
- ✅ **Inconsistent naming** - Does this match naming conventions used elsewhere?
- ✅ **Missing edge cases** - Based on similar code, what scenarios are unhandled?
- ✅ **Code duplication** - Could this reuse existing utilities?
- ✅ **API design issues** - Is this interface consistent with the rest of the API?
- ✅ **Documentation gaps** - Are complex sections unexplained?
- ✅ **Test coverage gaps** - Are critical paths untested?

**Tools**:
- **Claude Code GitHub Action** - Automated PR reviews using Claude
- **OpenAI GPT-4 Actions** - Similar capability with GPT-4
- **Custom review bots** - Self-hosted LLM review pipelines

**Example: Claude Code GitHub Action**

Add to `.github/workflows/claude-review.yml`:

```yaml
name: Claude Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  claude-review:
    runs-on: ubuntu-latest
    # Only run after CI passes
    needs: [lint, test, build]

    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Fetch full history for context

      - name: Claude Code Review
        uses: anthropics/claude-code-review-action@v1
        with:
          anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
          # Provide codebase context
          include_files: |
            CLAUDE.md
            ARCHITECTURE.md
            docs/**/*.md
          # Focused review prompts
          review_focus: |
            - Does this follow existing patterns in the codebase?
            - Is this the simplest solution to the problem?
            - Are there obvious edge cases that aren't handled?
            - Is the naming consistent with similar code?
            - Could this reuse existing utilities instead of duplicating code?
          # Make it non-blocking but visible
          post_comments: true
          require_approval: false
```

**Example: Custom Review Prompt**

You can customize what the LLM focuses on:

```yaml
review_focus: |
  Review this PR with the following criteria:

  1. **Simplicity**: Is this the simplest approach? Flag unnecessary abstractions.
  2. **Patterns**: Compare with files in src/services/* - does this follow the same patterns?
  3. **Error handling**: Check against CLAUDE.md guidelines - is error handling consistent?
  4. **Testing**: Are the test cases comprehensive compared to similar features?
  5. **Performance**: Are there obvious performance issues (N+1 queries, missing indexes)?
  6. **Security**: Any potential security issues (SQL injection, XSS, auth bypass)?

  Be specific in feedback - reference file names and line numbers.
  If code looks good, say so explicitly.
```

**Benefits over deterministic checks**:
- **Context-aware**: Understands project-specific patterns that aren't easily encoded in linter rules
- **Simplicity enforcement**: Can identify overcomplicated solutions that technically work but are unnecessarily complex
- **Learning from codebase**: Compares new code against existing examples to ensure consistency
- **Natural language feedback**: Explains *why* something is an issue, not just *that* it's an issue

**Benefits over human review**:
- **Instant feedback**: Runs in minutes, not hours/days
- **Always available**: No need to wait for reviewer availability
- **Consistent**: Applies the same standards every time
- **Scales**: Can review unlimited PRs without reviewer fatigue
- **Filters issues**: Catches simple problems so humans can focus on complex decisions

**What LLM review should NOT replace**:
- ❌ Deterministic checks (linting, type checking) - use CI for these
- ❌ Critical security reviews - humans should verify security-sensitive changes
- ❌ Architectural decisions - humans better understand business context
- ❌ Final approval - humans should have final say

**Configuration strategies**:

1. **Informational mode** (recommended to start):
   - LLM posts review comments
   - Does not block merging
   - Team learns what it catches

2. **Advisory mode**:
   - LLM review required but can be dismissed
   - Human can override if LLM is wrong
   - Good for established projects

3. **Blocking mode** (use carefully):
   - PR cannot merge without LLM approval
   - Only for mature projects with well-defined patterns
   - Requires escape hatch for emergencies

**Example LLM review output**:

```
🤖 Claude Code Review

## ✅ Looks Good
- Error handling follows the pattern from `src/services/auth.go`
- Test coverage is comprehensive (95% of new code)
- Naming is consistent with existing API endpoints

## ⚠️ Suggestions

### Potential Simplification (src/handlers/user.go:45-67)
This validation logic could reuse the existing `validateUserInput()`
function from `src/utils/validation.go:23` instead of duplicating it.

### Missing Edge Case (src/handlers/user.go:78)
Similar endpoints like `CreatePost()` handle the case where the user
is soft-deleted. Should this endpoint check `user.DeletedAt`?

### Pattern Inconsistency (src/handlers/user.go:89)
Other handlers in this package return errors using `handleError()`
helper. This directly returns `http.Error()`. Suggest using the helper
for consistency.

---
Overall: Good PR! The suggestions above are minor - the core logic is solid.
```

**Integration with other layers**:

```
Layer 3 (CI/CD)         ──> Deterministic checks pass
         │
         ├──> All tests green ✅
         ├──> Linting clean ✅
         ├──> Build succeeds ✅
         │
         ▼
Layer 3.5 (LLM Review)  ──> Contextual review
         │
         ├──> Checks against codebase patterns
         ├──> Validates simplicity
         ├──> Suggests improvements
         │
         ▼
Layer 4 (Human Review)  ──> Final judgment
         │
         ├──> Architecture decisions
         ├──> Business logic verification
         └──> Approval to merge
```

**Key principle**: Each layer filters out increasingly subtle issues, so humans can focus their expertise where it matters most - on architectural decisions, business logic correctness, and strategic tradeoffs that require domain knowledge.

---

#### Layer 4: Human Review

**Purpose**: Final judgment on issues that require human expertise and business context.

**Characteristics**:
- **Speed**: Hours to days
- **Blocking**: YES (requires approval)
- **Scope**: High-level concerns

**What humans review**:
- ✅ **Architecture decisions** - Is this the right approach?
- ✅ **Logic correctness** - Does the algorithm make sense?
- ✅ **Edge cases** - Are all scenarios handled?
- ✅ **User experience** - Is the UI/API intuitive?
- ✅ **Security implications** - Could this be exploited?
- ✅ **Performance concerns** - Will this scale?
- ✅ **Maintainability** - Is this code readable?
- ✅ **Documentation** - Are comments/docs adequate?
- ✅ **Test coverage** - Are the right scenarios tested?
- ✅ **Naming** - Are names clear and consistent?

**PR Review Checklist** (add to `PULL_REQUEST_TEMPLATE.md`):

```markdown
## Pre-merge Checklist

### Automated Checks
- [ ] All CI checks passing
- [ ] Code coverage maintained or improved
- [ ] No security vulnerabilities introduced

### Human Review
- [ ] Architecture reviewed and approved
- [ ] Logic correctness verified
- [ ] Edge cases identified and handled
- [ ] User experience considered
- [ ] Performance implications assessed
- [ ] Code is maintainable and well-documented
- [ ] Tests cover critical paths
- [ ] No obvious security issues
```

---

#### How the Layers Work Together

**Example workflow when Claude makes a change**:

1. **Layer 1 (< 1s)**: Hook formats the file, Claude sees quick feedback
   - ⚠️ "Line too long" → Claude can ignore or fix

2. **Layer 2 (< 30s)**: Developer commits, pre-commit runs full lint
   - ❌ "Type error on line 45" → Commit blocked, must fix

3. **Layer 3 (1-10 min)**: Developer pushes, CI runs comprehensive checks
   - ❌ "Integration test failed" → Must fix before merge

4. **Layer 3.5 (1-5 min)**: LLM reviews the PR after CI passes
   - 💡 "This could reuse the existing `validateUser()` utility" → Suggestion for simplification

5. **Layer 4 (hours/days)**: Human reviewer examines the PR
   - 💬 "This approach won't scale, consider caching" → Architecture feedback

**Each layer catches different issues**:

| Issue Type | Layer 1 | Layer 2 | Layer 3 | Layer 3.5 | Layer 4 |
|------------|---------|---------|---------|-----------|---------|
| Formatting | ✅ | ✅ | ✅ | - | - |
| Syntax errors | ⚠️ | ✅ | ✅ | - | - |
| Type errors | - | ✅ | ✅ | - | - |
| Failing tests | - | ⚠️ | ✅ | - | - |
| Security vulns | - | ⚠️ | ✅ | ⚠️ | ✅ |
| Logic errors | - | - | ⚠️ | ⚠️ | ✅ |
| Pattern violations | - | - | - | ✅ | ✅ |
| Unnecessary complexity | - | - | - | ✅ | ✅ |
| Architecture issues | - | - | - | ⚠️ | ✅ |

**Legend**: ✅ Catches reliably | ⚠️ Sometimes catches | - Doesn't check

---

#### Practical Configuration Strategy

**Start simple, add layers as needed**:

1. **Minimum viable setup** (for solo projects):
   - Layer 1: Fast hooks (formatting only)
   - Layer 2: Pre-commit with basic linting
   - Layer 3: CI with tests
   - Layer 4: Self-review checklist

2. **Team projects** (small to medium):
   - Layer 1: Fast hooks
   - Layer 2: Pre-commit (comprehensive)
   - Layer 3: CI with full test suite
   - Layer 3.5: Optional LLM review (informational mode)
   - Layer 4: Required human PR approvals

3. **Team projects** (large scale):
   - All 5 layers fully configured
   - Layer 3.5: LLM review in advisory mode (can be dismissed)
   - Protected branches with required reviews
   - CODEOWNERS for critical paths

4. **Open source**:
   - All 5 layers
   - Layer 3.5: LLM review helps maintainers triage
   - Multiple CI providers (GitHub Actions + CircleCI)
   - Extensive documentation requirements
   - Community contribution guidelines

**Keep Layer 1 fast**:
- ⏱️ Target: < 1 second total
- 🎯 Focus: Auto-fixable issues only
- ⚠️ Non-blocking: Never fail, only warn
- 🚀 Benefit: Claude gets instant feedback without disruption

**Make Layer 2 comprehensive**:
- ⏱️ Target: < 30 seconds total
- 🎯 Focus: Everything that should never reach main branch
- 🚫 Blocking: Prevent commits that would fail CI
- 💰 Benefit: Save CI time and cost

**Make Layer 3 thorough**:
- ⏱️ Target: < 10 minutes total
- 🎯 Focus: Everything automated
- 🔒 Blocking: Gate to production
- 🏆 Benefit: Confidence before merge

---

### 3. Project Guidelines with CLAUDE.md

**Problem**: Claude Code doesn't know project-specific conventions, common pitfalls, or the preferred development workflow.

**Solution**: Create a `CLAUDE.md` file in your project root with clear guidelines.

#### Template

```markdown
# Code Guidelines for Claude

## [Language] Code Guidelines

### Error Handling

[Describe common error handling patterns and anti-patterns in your codebase]

**❌ Wrong:**
\`\`\`[language]
[example of incorrect pattern]
\`\`\`

**✅ Correct:**
\`\`\`[language]
[example of correct pattern]
\`\`\`

### Code Style

- [List important style conventions]
- [Mention any automated formatters/linters]
- [Highlight common mistakes to avoid]

### Testing

- [Where tests should be placed]
- [How to run tests]
- [Required test coverage expectations]

## Development Process

When making changes to this codebase, follow these steps:

### 1. Start the Development Server

Before making any changes, start the auto-reloading development server:

\`\`\`bash
make dev
\`\`\`

This command runs the application with automatic reloading on file changes.

### 2. Make Changes Iteratively

- Keep changes focused and minimal
- Make one logical change at a time
- Follow existing code patterns and conventions

### 3. Verify Your Changes

After making changes, verify they work correctly:

- **Tail the logs** to see server output and debug issues:
  \`\`\`bash
  make dev-logs
  \`\`\`
- **Test the changes** by making API requests or interacting with the UI
- **Check for errors** in the log output
- Test both success and error cases

**Example workflow:**
\`\`\`bash
# Terminal 1: View logs
make dev-logs

# Terminal 2: Make API requests to test
curl http://localhost:8080/api/endpoint

# Or use a GUI tool like Postman, Insomnia, etc.
\`\`\`

### 4. Run Tests

Before committing, ensure all tests pass:

\`\`\`bash
make test
\`\`\`

## Common Pitfalls

### [Pitfall 1]
[Description and how to avoid]

### [Pitfall 2]
[Description and how to avoid]

## Architecture Overview

[Brief description of project structure]

\`\`\`
project-root/
├── src/           # [Description]
├── tests/         # [Description]
├── config/        # [Description]
└── ...
\`\`\`

## Additional Resources

- [Link to API documentation]
- [Link to architecture decision records]
- [Link to deployment guide]
```

**Example for Different Stacks**:

<details>
<summary>Python/Django Example</summary>

```markdown
# Code Guidelines for Claude

## Python Code Guidelines

### Error Handling

Always use explicit exception handling. Never use bare `except:` clauses.

**❌ Wrong:**
```python
try:
    result = api_call()
except:
    pass
```

**✅ Correct:**
```python
try:
    result = api_call()
except APIError as e:
    logger.error(f"API call failed: {e}")
    raise
```

### Development Process

1. Start dev server: `make dev`
2. Run migrations: `python manage.py migrate`
3. Tail logs: `make dev-logs`
4. Run tests: `pytest`
```
</details>

<details>
<summary>Node.js/Express Example</summary>

```markdown
# Code Guidelines for Claude

## Node.js Code Guidelines

### Error Handling

Always handle promise rejections and pass errors to the next middleware.

**❌ Wrong:**
```javascript
app.get('/users', async (req, res) => {
  const users = await db.getUsers()
  res.json(users)
})
```

**✅ Correct:**
```javascript
app.get('/users', async (req, res, next) => {
  try {
    const users = await db.getUsers()
    res.json(users)
  } catch (error) {
    next(error)
  }
})
```

### Development Process

1. Start dev server: `npm run dev`
2. Tail logs: `make dev-logs`
3. Run tests: `npm test`
4. Run linter: `npm run lint`
```
</details>

**Key Benefits**:
- Reduces back-and-forth questions
- Claude follows project conventions automatically
- Documents tribal knowledge
- Onboards new developers (human or AI)

---

## Complete Setup Checklist

Use this checklist when setting up a new project:

- [ ] **Process Management**
  - [ ] Install process manager (Hivemind, Foreman, etc.)
  - [ ] Create `Procfile` with all services
  - [ ] Add `make dev`, `make dev-logs`, `make dev-stop` targets

- [ ] **Logging**
  - [ ] Configure unified log file (`dev-server.log`)
  - [ ] Add PID file tracking (`.dev-server.pid`)
  - [ ] Add log files to `.gitignore`
  - [ ] Test log tailing with ANSI stripping

- [ ] **Hooks**
  - [ ] Create `.claude/settings.json`
  - [ ] Create `.claude/hooks/session-start.sh`
  - [ ] Create `.claude/hooks/lint-file.sh`
  - [ ] Make hooks executable (`chmod +x`)
  - [ ] Add `lint-file` target to Makefile with language detection

- [ ] **Linters/Formatters**
  - [ ] Install formatters for your languages
  - [ ] Install linters for your languages
  - [ ] Configure linter rules (`.eslintrc`, `pylintrc`, etc.)
  - [ ] Add `format-all` and `lint-all` Makefile targets
  - [ ] Test that hooks trigger linting correctly

- [ ] **Guidelines**
  - [ ] Create `CLAUDE.md` with project-specific instructions
  - [ ] Document error handling patterns
  - [ ] Document development workflow
  - [ ] Document testing requirements
  - [ ] Include common pitfalls and solutions

- [ ] **Makefile**
  - [ ] Add self-documenting help target
  - [ ] Add install target for all dependencies
  - [ ] Add test target
  - [ ] Add lint/format targets
  - [ ] Ensure all targets work from project root

- [ ] **Layered Quality Checks**
  - [ ] **Layer 1 (Hooks)**: Already configured above
  - [ ] **Layer 2 (Pre-commit)**:
    - [ ] Install pre-commit framework or equivalent
    - [ ] Create `.pre-commit-config.yaml`
    - [ ] Configure hooks for linting, type checking, tests
    - [ ] Run `pre-commit install`
    - [ ] Test pre-commit with `pre-commit run --all-files`
  - [ ] **Layer 3 (CI/CD)**:
    - [ ] Create `.github/workflows/ci.yml` or equivalent
    - [ ] Configure lint, test, security, and build jobs
    - [ ] Set up branch protection rules
    - [ ] Verify CI runs on PRs
  - [ ] **Layer 3.5 (LLM Review)** (Optional):
    - [ ] Set up Claude Code GitHub Action or equivalent
    - [ ] Configure review focus prompts
    - [ ] Add ANTHROPIC_API_KEY secret
    - [ ] Test on a sample PR
  - [ ] **Layer 4 (Human Review)**:
    - [ ] Create PR template with review checklist
    - [ ] Configure required approvals
    - [ ] Set up CODEOWNERS if needed

- [ ] **Testing**
  - [ ] Test full workflow: `make dev` → edit file → check logs → verify linting
  - [ ] Test session start hook
  - [ ] Test lint-file hook with different file types
  - [ ] Verify dev server auto-starts correctly
  - [ ] Test pre-commit blocks bad commits
  - [ ] Verify CI runs and blocks bad PRs

---

## Minimal Makefile Template

```makefile
.PHONY: help install dev dev-logs dev-stop test lint-file format-all lint-all

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install all dependencies
	@echo "Installing dependencies..."
	# Add your install commands here

dev: ## Start dev server with unified logging
	@if [ -f .dev-server.pid ]; then \
		PID=$$(cat .dev-server.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Dev server already running (PID: $$PID)"; \
			exit 1; \
		fi; \
	fi
	@mkdir -p .logs
	@nohup hivemind Procfile > dev-server.log 2>&1 & echo $$! > .dev-server.pid
	@echo "Dev server started (PID: $$(cat .dev-server.pid))"

dev-logs: ## Tail dev server logs
	@tail -f dev-server.log | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g'

dev-stop: ## Stop dev server
	@if [ -f .dev-server.pid ]; then \
		kill $$(cat .dev-server.pid) && rm .dev-server.pid; \
		echo "Dev server stopped"; \
	fi

test: ## Run tests
	# Add your test command here

lint-file: ## Lint single file (FILE=path/to/file)
	@# Add file type detection and linting here

format-all: ## Format all code
	# Add formatting commands here

lint-all: ## Lint all code
	# Add linting commands here
```

---

## Example: Full Setup for a Node.js + React Project

<details>
<summary>Click to expand full example</summary>

### File: `Procfile`
```procfile
api: cd api && npm run dev
web: cd web && npm run dev
```

### File: `Makefile`
```makefile
.PHONY: install dev dev-logs dev-stop lint-file

install:
	cd api && npm install
	cd web && npm install
	npm install -g hivemind

dev:
	@if [ -f .dev-server.pid ]; then \
		PID=$$(cat .dev-server.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Dev server already running (PID: $$PID)"; \
			exit 1; \
		fi; \
	fi
	@nohup hivemind Procfile > dev-server.log 2>&1 & echo $$! > .dev-server.pid
	@echo "Dev server started"

dev-logs:
	@tail -f dev-server.log | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g'

dev-stop:
	@if [ -f .dev-server.pid ]; then \
		kill $$(cat .dev-server.pid) && rm .dev-server.pid; \
	fi

lint-file:
	@if echo "$(FILE)" | grep -qE '\.(ts|tsx|js|jsx)$$'; then \
		npx prettier --write "$(FILE)"; \
		npx eslint --fix "$(FILE)" || true; \
	fi
```

### File: `.claude/settings.json`
```json
{
  "hooks": {
    "SessionStart": [{"hooks": [{"type": "command", "command": "./.claude/hooks/session-start.sh"}]}],
    "PostToolUse": [{"matcher": "Edit|Write", "hooks": [{"type": "command", "command": "./.claude/hooks/lint-file.sh"}]}]
  }
}
```

### File: `.claude/hooks/session-start.sh`
```bash
#!/bin/bash
cd "$CLAUDE_PROJECT_DIR" || exit 1
echo "Installing dependencies..."
make install &>/dev/null && echo "✓ Dependencies installed"
echo "Starting dev server..."
make dev 2>&1 | grep -q "already running" && echo "✓ Server running" || echo "✓ Server started"
```

### File: `.claude/hooks/lint-file.sh`
```bash
#!/bin/bash
FILE=$(cat | jq -r ".tool_input.file_path")
cd "$(git rev-parse --show-toplevel)" || exit 1
LINT_OUTPUT=$(make lint-file FILE="$FILE" 2>&1)
if [ -n "$LINT_OUTPUT" ]; then
  jq -n --arg ctx "$LINT_OUTPUT" '{"hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": $ctx}}'
else
  echo "{}"
fi
```

### File: `CLAUDE.md`
```markdown
# Code Guidelines for Claude

## Development Process

1. **Start dev server**: `make dev`
2. **View logs**: `make dev-logs`
3. **Make changes** and check logs for errors
4. **Test** by visiting http://localhost:3000 (web) and http://localhost:4000 (api)

## Error Handling

Always use try-catch for async operations.

**❌ Wrong:**
\`\`\`javascript
const data = await fetch('/api/users')
\`\`\`

**✅ Correct:**
\`\`\`javascript
try {
  const data = await fetch('/api/users')
} catch (error) {
  console.error('Failed to fetch users:', error)
  throw error
}
\`\`\`
```

</details>

---

## Advanced: Alternative Approaches

### Using Docker Compose Instead of Hivemind

If you prefer containerization:

```yaml
# docker-compose.yml
version: '3.8'
services:
  backend:
    build: ./backend
    volumes:
      - ./backend:/app
    ports:
      - "8080:8080"
    command: [your dev command]

  frontend:
    build: ./frontend
    volumes:
      - ./frontend:/app
    ports:
      - "3000:3000"
    command: npm run dev
```

```makefile
dev:
	docker-compose up

dev-logs:
	docker-compose logs -f
```

### Using tmux for Terminal Multiplexing

```bash
# .claude/hooks/session-start.sh
tmux new-session -d -s dev "cd backend && npm run dev"
tmux split-window -h -t dev "cd frontend && npm run dev"
tmux attach -t dev
```

---

## Why This Works: The Feedback Loop

This setup creates a tight feedback loop:

```
┌─────────────────────────────────────────────────────┐
│ 1. Claude edits file                                │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│ 2. PostToolUse hook triggers lint-file.sh          │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│ 3. Linter runs and returns errors as context       │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│ 4. Claude sees errors and fixes them immediately   │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│ 5. Dev server auto-reloads with changes            │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│ 6. Claude checks logs via make dev-logs            │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│ 7. If errors in logs, Claude fixes and repeats     │
└─────────────────────────────────────────────────────┘
```

**Time from error introduction to detection**: < 1 second (via hooks)
**Time from detection to fix**: < 30 seconds (Claude re-edits)

Compare this to manual detection:
- No linting: Errors discovered at compile/runtime (minutes to hours later)
- Manual linting: Developer runs linter periodically (minutes later)
- CI/CD only: Errors discovered after commit (minutes to hours later)

---

## Troubleshooting

### Hooks Not Running

1. Check hook files are executable: `chmod +x .claude/hooks/*.sh`
2. Verify `.claude/settings.json` syntax with `jq . .claude/settings.json`
3. Check Claude Code output for hook errors

### Dev Server Won't Start

1. Check if port is already in use: `lsof -i :8080`
2. Verify Procfile syntax
3. Test commands individually: `go run ./cmd/server/main.go`
4. Check for stale PID file: `rm .dev-server.pid`

### Linter Not Working

1. Verify linter is installed: `which eslint` or `which pylint`
2. Test Makefile target manually: `make lint-file FILE=src/main.js`
3. Check file extension matching in Makefile
4. Ensure linter exits with code 0 (use `|| true` to prevent failures)

---

## Summary

By implementing these four components:

1. **Unified logging** → See everything in one place
2. **Automatic linting** → Catch errors immediately
3. **Auto-starting dev server** → Begin working instantly
4. **Clear guidelines** → Reduce ambiguity

You create an environment where Claude Code can:
- Detect its own mistakes within seconds
- Fix errors before they accumulate
- Follow project conventions automatically
- Work iteratively with rapid feedback

This results in:
- Fewer bugs in final code
- Less manual intervention needed
- Faster development cycles
- More consistent code quality

**Start with the basics** (Procfile + Makefile), **add hooks** for linting, and **document guidelines** in CLAUDE.md. You'll have a robust development environment that works with any language or framework.
