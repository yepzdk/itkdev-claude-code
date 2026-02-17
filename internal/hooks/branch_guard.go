package hooks

import (
	"encoding/json"
	"os/exec"
	"strings"
)

func init() {
	Register("branch-guard", branchGuardHook)
}

// branchGuardHook enforces the branch-based PR workflow. It handles two events:
//   - SessionStart: injects a reminder about the branching workflow if on main
//   - PreToolUse (Bash): blocks git commit/push operations when on the main branch
func branchGuardHook(input *Input) error {
	switch input.HookEventName {
	case "SessionStart":
		return branchGuardSessionStart(input)
	case "PreToolUse":
		return branchGuardPreToolUse(input)
	default:
		ExitOK()
		return nil
	}
}

func branchGuardSessionStart(input *Input) error {
	branch := currentBranch(input.Cwd)
	if branch == "main" || branch == "master" {
		WriteOutput(&Output{
			HookSpecific: &HookSpecificOuput{
				HookEventName:     "SessionStart",
				AdditionalContext: "You are currently on the `" + branch + "` branch. Do NOT commit or push directly to this branch. Create a feature branch first (e.g., `git checkout -b feat/my-feature`), then open a PR when ready.",
			},
		})
		return nil
	}
	ExitOK()
	return nil
}

func branchGuardPreToolUse(input *Input) error {
	if input.ToolName != "Bash" {
		ExitOK()
		return nil
	}

	var bash BashToolInput
	if err := json.Unmarshal(input.ToolInput, &bash); err != nil {
		ExitOK()
		return nil
	}

	cmd := strings.TrimSpace(bash.Command)
	if !isGitCommand(cmd) {
		ExitOK()
		return nil
	}

	// Check for push targeting main regardless of current branch
	if isPushToMain(cmd) {
		BlockWithError("Blocked: Do not push directly to main. Push your feature branch and open a PR instead.\n\nExample:\n  git push -u origin feat/my-feature\n  gh pr create")
		return nil
	}

	// For commit and push-without-explicit-target, check current branch
	branch := currentBranch(input.Cwd)
	if branch != "main" && branch != "master" {
		ExitOK()
		return nil
	}

	if isGitCommit(cmd) {
		BlockWithError("Blocked: Do not commit directly to " + branch + ". Create a feature branch first.\n\nExample:\n  git checkout -b feat/my-feature")
		return nil
	}

	if isGitPush(cmd) {
		BlockWithError("Blocked: Do not push directly to " + branch + ". Push your feature branch and open a PR instead.\n\nExample:\n  git checkout -b feat/my-feature\n  git push -u origin feat/my-feature\n  gh pr create")
		return nil
	}

	ExitOK()
	return nil
}

// currentBranch returns the current git branch name for the given directory.
// Returns empty string if git is not available or the directory is not a repo.
func currentBranch(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if cwd != "" {
		cmd.Dir = cwd
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// isGitCommand checks if a command string starts with or contains a git command.
func isGitCommand(cmd string) bool {
	// Handle chained commands: "git add . && git commit -m ..."
	parts := splitChainedCommands(cmd)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(trimmed, "git ") || trimmed == "git" {
			return true
		}
	}
	return false
}

// isGitCommit checks if the command contains a git commit operation.
func isGitCommit(cmd string) bool {
	parts := splitChainedCommands(cmd)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(trimmed, "git commit") {
			return true
		}
	}
	return false
}

// isGitPush checks if the command contains a git push operation.
func isGitPush(cmd string) bool {
	parts := splitChainedCommands(cmd)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(trimmed, "git push") {
			return true
		}
	}
	return false
}

// isPushToMain checks if the command explicitly pushes to main/master.
func isPushToMain(cmd string) bool {
	parts := splitChainedCommands(cmd)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if !strings.HasPrefix(trimmed, "git push") {
			continue
		}
		// Check for "git push origin main" or "git push origin master"
		words := strings.Fields(trimmed)
		for i, w := range words {
			if i > 1 && (w == "main" || w == "master") {
				return true
			}
		}
	}
	return false
}

// splitChainedCommands splits a shell command string on && and ; delimiters.
func splitChainedCommands(cmd string) []string {
	// Split on && first, then on ;
	var results []string
	for _, part := range strings.Split(cmd, "&&") {
		for _, sub := range strings.Split(part, ";") {
			results = append(results, sub)
		}
	}
	return results
}
