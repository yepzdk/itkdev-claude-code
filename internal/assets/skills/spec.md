# /spec — Spec-Driven Development

Structured plan, implement, and verify workflow for complex tasks.

## Usage

```
/spec <task description>
```

Everything after `/spec` is the task description.

## How It Works

`/spec` is a dispatcher that routes to phase skills based on the plan file status:

| Plan Status | Action |
|-------------|--------|
| No plan exists | Create plan → `spec-plan` |
| PENDING, not approved | Continue planning → `spec-plan` |
| PENDING, approved | Implement → `spec-implement` |
| COMPLETE | Verify → `spec-verify` |
| VERIFIED | Done — report completion |

## Plan File

Plans are stored in `docs/plans/YYYY-MM-DD-<slug>.md` with this header:

```markdown
# <Title>

Status: PENDING | COMPLETE | VERIFIED
Approved: Yes | No

## Tasks

- [ ] Task 1
- [ ] Task 2
```

## Phase Flow

```
spec-plan → (user approves) → spec-implement → spec-verify
                                    ↑                |
                                    └── (issues) ────┘
```

## Dispatch Logic

1. Check for existing plan files in `docs/plans/`
2. Read the most recent plan's Status and Approved fields
3. Invoke the appropriate phase skill via `Skill()`
4. Each phase updates the plan status when complete

## Continuing an Existing Plan

```
/spec --continue docs/plans/2026-02-17-my-feature.md
```

If `--continue` is provided with a plan path, skip plan detection and use that file directly.
