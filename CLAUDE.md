# Code Guidelines for Claude

## Go Code Guidelines

### Error Handling

Always check and handle error returns from function calls. The linter (`golangci-lint`) enforces this with the `errcheck` rule.

**❌ Wrong:**
```go
repo.Update(1, models.UpdateTodoRequest{Completed: &completed})
```

**✅ Correct:**
```go
_, err := repo.Update(1, models.UpdateTodoRequest{Completed: &completed})
if err != nil {
    // Handle error appropriately (return it, log it, or fail the test)
    t.Fatalf("Failed to update todo: %v", err)
}
```

- When calling functions that return errors, always assign the error value
- If you intentionally want to ignore an error, use `_ = ` explicitly
- In tests, use `t.Fatalf()` or `t.Errorf()` to properly fail/report the test
- In production code, return the error up the call stack or handle it appropriately

This ensures no errors are silently ignored and maintains code reliability.

## Development Process

When making changes to this codebase, follow these steps:

### 1. Start the Development Server

Before making any changes, start the auto-reloading development server:

```bash
make dev
```

This command runs the application with automatic reloading on file changes, allowing you to see your changes in real-time.

### 2. Make Changes Simply

- Keep changes focused and minimal
- Make one logical change at a time
- Follow the existing code patterns and conventions
- Ensure all error handling is in place (see Error Handling guidelines above)

### 3. Check Your Changes

After making changes, verify they work correctly:

- **Make API requests** to test the endpoints you modified
- **Tail the logs** to see server output and debug any issues:
  ```bash
  make dev-logs
  ```
- Test both success and error cases
- Ensure the application still compiles and runs without errors

**Example workflow:**
```bash
# Terminal 1: Start dev server
make dev

# Terminal 2: Tail logs
make dev-logs

# Terminal 3: Make API requests to test
curl http://localhost:8080/todos
```

This iterative process ensures changes work correctly before committing.
