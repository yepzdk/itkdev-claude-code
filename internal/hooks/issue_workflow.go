package hooks

import (
	"os"
	"regexp"
	"strings"

	"github.com/itk-dev/itkdev-claude-code/internal/config"
)

func init() {
	Register("issue-workflow", issueWorkflowHook)
}

// issuePattern matches common ways users refer to working on GitHub issues.
// Examples: "issue #13", "issue 13", "work on #42", "fix #7", "tackle #99"
var issuePattern = regexp.MustCompile(`(?i)(?:issue\s*#?\d+|#\d+\b|work\s+on\s+(?:issue\s+)?#?\d+|fix\s+#?\d+|tackle\s+#?\d+|close\s+#?\d+|resolve\s+#?\d+)`)

// issueWorkflowPlugin is the required plugin for the issue workflow skill.
var issueWorkflowPlugin = config.RequiredPlugin{
	Name:        "itkdev-tools",
	Marketplace: "itkdev-marketplace",
}

// issueWorkflowHook enforces the itkdev-issue-workflow skill when users work
// on GitHub issues. It handles two events:
//   - UserPromptSubmit: detects issue-related prompts and injects a directive
//     to use the itkdev-issue-workflow skill
//   - SessionStart: checks for ICC_ISSUE_ID env var and injects issue context
func issueWorkflowHook(input *Input) error {
	switch input.HookEventName {
	case "UserPromptSubmit":
		return issueWorkflowPromptSubmit(input)
	case "SessionStart":
		return issueWorkflowSessionStart(input)
	default:
		ExitOK()
		return nil
	}
}

// issueWorkflowPromptSubmit checks if the user's prompt mentions a GitHub issue
// and injects guidance to use the itkdev-issue-workflow skill.
func issueWorkflowPromptSubmit(input *Input) error {
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		ExitOK()
		return nil
	}

	if !issuePattern.MatchString(prompt) {
		ExitOK()
		return nil
	}

	if msg := pluginMissingMessage(); msg != "" {
		WriteOutput(&Output{
			HookSpecific: &HookSpecificOuput{
				HookEventName:     "UserPromptSubmit",
				AdditionalContext: msg,
			},
		})
		return nil
	}

	WriteOutput(&Output{
		HookSpecific: &HookSpecificOuput{
			HookEventName: "UserPromptSubmit",
			AdditionalContext: "The user wants to work on a GitHub issue. " +
				"You MUST use the itkdev-issue-workflow skill to handle this request. " +
				"Invoke it with: Skill(itkdev-tools:itkdev-issue-workflow). " +
				"Do NOT work on the issue manually — always delegate to the skill.",
		},
	})
	return nil
}

// issueWorkflowSessionStart checks if ICC_ISSUE_ID is set and injects
// issue context at session start.
func issueWorkflowSessionStart(_ *Input) error {
	issueID := os.Getenv(config.EnvPrefix + "_ISSUE_ID")
	if issueID == "" {
		ExitOK()
		return nil
	}

	if msg := pluginMissingMessage(); msg != "" {
		WriteOutput(&Output{
			HookSpecific: &HookSpecificOuput{
				HookEventName:     "SessionStart",
				AdditionalContext: msg,
			},
		})
		return nil
	}

	WriteOutput(&Output{
		HookSpecific: &HookSpecificOuput{
			HookEventName: "SessionStart",
			AdditionalContext: "This session was started to work on GitHub issue #" + issueID + ". " +
				"You MUST use the itkdev-issue-workflow skill to handle this issue. " +
				"Invoke it with: Skill(itkdev-tools:itkdev-issue-workflow). " +
				"Pass the issue number " + issueID + " when the skill asks for it. " +
				"Do NOT work on the issue manually — always delegate to the skill.",
		},
	})
	return nil
}

// pluginMissingMessage returns an instruction message if the required plugin
// is not installed. Returns empty string if the plugin is available.
func pluginMissingMessage() string {
	installed, err := config.IsPluginInstalled(issueWorkflowPlugin)
	if err != nil {
		// Can't determine status — assume available to avoid blocking work.
		return ""
	}
	if installed {
		return ""
	}
	return "The itkdev-issue-workflow skill requires the itkdev-tools plugin, " +
		"which is not currently installed. " +
		"Tell the user to install it in Claude Code with:\n" +
		"  /plugin marketplace add itk-dev/itkdev-claude-plugins\n" +
		"  /plugin install itkdev-tools@itkdev-marketplace\n" +
		"Then retry this request."
}
