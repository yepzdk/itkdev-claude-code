# spec-verify — Verification Phase

Verify that the implementation meets the plan's requirements.

## Steps

### 1. Run Tests

Run the project's full test suite. All tests must pass.

```
If tests fail → fix the failures, re-run until green.
```

### 2. Run the Program

If there's a runnable entry point, execute it and verify real output. Tests passing with mocks does not prove the program works.

### 3. Review Against Plan

Read each task in the plan and verify it was implemented correctly:

- Does the code match what was planned?
- Are edge cases handled?
- Are there any tasks that were skipped?

### 4. Check Code Quality

- Run the project's linter/formatter
- Verify no files exceed 300 lines (500 hard limit)
- Check for unused imports, dead code, or obvious issues

### 5. Update Status

If everything passes:

```markdown
Status: VERIFIED
```

If issues are found:

1. Document what needs fixing
2. Set status back to PENDING:
   ```markdown
   Status: PENDING
   ```
3. Invoke `Skill('spec-implement')` to fix the issues

## The Feedback Loop

```
spec-verify finds issues → Status: PENDING → spec-implement fixes → COMPLETE → spec-verify → ... → VERIFIED
```

Repeat until clean.
