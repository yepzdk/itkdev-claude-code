# spec-plan — Planning Phase

Create a structured implementation plan for user approval.

## Steps

### 1. Explore the Codebase

Understand existing patterns, architecture, and relevant files before designing a solution.

- Read related source files
- Identify existing patterns to follow
- Note dependencies and constraints

### 2. Design the Plan

Create `docs/plans/YYYY-MM-DD-<slug>.md` with:

```markdown
# <Title>

Status: PENDING
Approved: No

## Summary

<Brief description of what will be built and why>

## Tasks

- [ ] Task 1: <description>
- [ ] Task 2: <description>
- [ ] Task 3: <description>

Progress: Done 0 / Left N / Total N
```

**Task guidelines:**
- Each task should be independently testable
- Order tasks by dependency (earlier tasks don't depend on later ones)
- Include test tasks where appropriate
- Keep tasks small and focused

### 3. Present for Approval

Show the plan to the user and wait for explicit approval. Do not proceed until the user says to go ahead.

After approval, update the plan:

```markdown
Approved: Yes
```

Then invoke `Skill('spec-implement')` to begin implementation.
