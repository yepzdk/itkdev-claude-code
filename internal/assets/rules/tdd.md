## Test-Driven Development

Write a failing test before writing production code. Follow the RED-GREEN-REFACTOR cycle.

### The Cycle

1. **RED** — Write one minimal test for the desired behavior. Run it. Verify it fails for the right reason (missing feature, not a syntax error).
2. **GREEN** — Write the simplest code that makes the test pass. Nothing more.
3. **REFACTOR** — Improve code quality while keeping tests green.

### Rules

- **Always verify RED** — Run the test and confirm it fails before writing production code. A test that passes immediately proves nothing.
- **Always verify GREEN** — Run the test and confirm it passes. Don't assume.
- **Minimal code** — Implement only what the test requires. No extra features.
- **No refactoring during GREEN** — Get to green first, then refactor.
- **Real code over mocks** — Only mock external dependencies (APIs, databases, file systems). Don't mock your own code.

### When TDD Applies

- New functions, methods, API endpoints, business logic
- Bug fixes — write a test that reproduces the bug first

### When TDD Doesn't Apply

- Documentation-only changes
- Configuration file updates
- Formatting/style-only changes

### Recovery

If you wrote production code before the test: write the test now (it will pass — that's fine in recovery mode). Verify it tests the right behavior and would catch regressions. Apply TDD properly for the remaining work.
