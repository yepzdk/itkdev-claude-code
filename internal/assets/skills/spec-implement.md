# spec-implement — Implementation Phase

Implement each task from the approved plan using TDD.

## Steps

### 1. Read the Plan

Load the plan file and identify the first uncompleted task (`- [ ]`).

### 2. Implement Each Task

For each task, follow the TDD cycle:

1. **Write a failing test** for the task's behavior
2. **Run the test** — verify it fails for the right reason
3. **Write minimal code** to make the test pass
4. **Run the test** — verify it passes
5. **Refactor** if needed, keeping tests green

### 3. Update the Plan

After completing each task:

1. Mark it done: `- [ ]` → `- [x]`
2. Update the progress line: increment Done, decrement Left

### 4. Mark Complete

After all tasks are done, update the plan status:

```markdown
Status: COMPLETE
```

Then invoke `Skill('spec-verify')` to begin verification.

## Guidelines

- Follow the plan's task order — earlier tasks may be dependencies for later ones
- If you discover work not covered by the plan, add it as a new task before implementing
- If a task requires a significant architectural change not in the plan, stop and ask the user
- Run the full test suite periodically to catch regressions
