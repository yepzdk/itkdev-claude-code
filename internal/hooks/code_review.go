package hooks

import (
	"os"
	"regexp"
	"strings"

	"github.com/itk-dev/itkdev-claude-code/internal/config"
)

func init() {
	Register("code-review", codeReviewHook)
}

// codeReviewPattern matches common ways users ask for code reviews.
// Examples: "review my code", "code review", "review this PR", "PR review",
// "check my changes", "look at this PR", "examine the code"
var codeReviewPattern = regexp.MustCompile(
	`(?i)(?:review(?:\s+(?:my|the|this)\s+)?(?:code|changes|pull\s*request|pr)|` +
		`code\s*review|` +
		`(?:pull\s*request|pr)\s*review|` +
		`(?:check|look\s+at|examine)\s+(?:(?:this|my|the)\s+)?(?:code|changes|pr))`)

// codeReviewPlugin is the required plugin for the code review skill.
var codeReviewPlugin = config.RequiredPlugin{
	Name:        "itkdev-tools",
	Marketplace: "itkdev-marketplace",
}

// codeReviewHook enforces the itkdev-code-review skill when users request
// code reviews. It handles two events:
//   - UserPromptSubmit: detects review-related prompts and injects a directive
//     to use the itkdev-code-review skill
//   - SessionStart: checks for ICC_REVIEW_ID env var and injects review context
func codeReviewHook(input *Input) error {
	switch input.HookEventName {
	case "UserPromptSubmit":
		return codeReviewPromptSubmit(input)
	case "SessionStart":
		return codeReviewSessionStart(input)
	default:
		ExitOK()
		return nil
	}
}

// codeReviewPromptSubmit checks if the user's prompt mentions a code review
// and injects guidance to use the itkdev-code-review skill.
func codeReviewPromptSubmit(input *Input) error {
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		ExitOK()
		return nil
	}

	if !codeReviewPattern.MatchString(prompt) {
		ExitOK()
		return nil
	}

	if msg := codeReviewPluginMissing(); msg != "" {
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
			AdditionalContext: "The user wants a code review. " +
				"You MUST use the itkdev-code-review skill to handle this request. " +
				"Invoke it with: Skill(itkdev-tools:itkdev-code-review). " +
				"Do NOT perform a manual code review — always delegate to the skill.",
		},
	})
	return nil
}

// codeReviewSessionStart checks if ICC_REVIEW_ID is set and injects
// review context at session start.
func codeReviewSessionStart(_ *Input) error {
	reviewID := os.Getenv(config.EnvPrefix + "_REVIEW_ID")
	if reviewID == "" {
		ExitOK()
		return nil
	}

	if msg := codeReviewPluginMissing(); msg != "" {
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
			AdditionalContext: "This session was started to review " + reviewID + ". " +
				"You MUST use the itkdev-code-review skill to perform the review. " +
				"Invoke it with: Skill(itkdev-tools:itkdev-code-review). " +
				"The target to review is: " + reviewID + ". " +
				"Do NOT perform a manual code review — always delegate to the skill.",
		},
	})
	return nil
}

// codeReviewPluginMissing returns an instruction message if the required plugin
// is not installed. Returns empty string if the plugin is available.
func codeReviewPluginMissing() string {
	installed, err := config.IsPluginInstalled(codeReviewPlugin)
	if err != nil {
		// Can't determine status — assume available to avoid blocking work.
		return ""
	}
	if installed {
		return ""
	}
	return "The itkdev-code-review skill requires the itkdev-tools plugin, " +
		"which is not currently installed. " +
		"Tell the user to install it in Claude Code with:\n" +
		"  /plugin marketplace add itk-dev/itkdev-claude-plugins\n" +
		"  /plugin install itkdev-tools@itkdev-marketplace\n" +
		"Then retry this request."
}
