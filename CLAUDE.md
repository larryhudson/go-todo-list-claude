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
