## Coding Standards

### Priority Order

**Correctness > Maintainability > Performance > Brevity**

### Core Principles

- **DRY** — Extract duplicated logic into reusable functions. If you're about to copy-paste, stop and create a shared function.
- **YAGNI** — Build only what's explicitly required. No abstractions for hypothetical future needs.
- **Single Responsibility** — Each function does one thing. If you need "and" to describe it, split it.

### Naming

Use descriptive names that reveal intent:
- Functions: `calculate_discount`, `validate_email`, `fetch_active_users`
- Avoid: `process`, `handle`, `data`, `temp`, `x`, `do_stuff`

### Code Organization

- **Imports** — Standard library, third-party, local. Remove unused imports immediately.
- **Dead code** — Delete it. Version control exists for a reason.
- **File size** — Keep production files under 300 lines. Above 500 is a hard limit — split immediately. Test files are exempt.

### Self-Correction

Fix obvious mistakes (syntax errors, typos, missing imports) immediately without asking. Reserve user communication for decisions, not minor fixes.
