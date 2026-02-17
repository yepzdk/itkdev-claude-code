## Git Operations

You may read git state freely. Never execute git write commands without explicit user permission.

### Always Allowed

Reading repository state — use these freely:

```
git status, git diff, git log, git show, git branch
```

File modifications (creating, editing, deleting files) are always allowed. This rule is about git commands, not file operations.

### Requires Explicit Permission

These commands need the user to say "commit", "push", etc.:

```
git add, git commit, git push, git pull, git fetch
git merge, git rebase, git reset, git revert, git stash
git checkout, git switch, git restore, git cherry-pick
```

### Key Rules

- **"Fix this bug" does NOT mean "commit it."** Make the fix, verify it works, then wait for the user to say commit.
- **Never use `git add -f`** to bypass `.gitignore`. If a file is ignored, tell the user and ask if they want to update `.gitignore`.
- **Commit ALL staged changes.** Don't selectively unstage files you think are unrelated — the user staged them intentionally.
- **Never amend** unless the user explicitly asks. After a hook failure, create a new commit.
