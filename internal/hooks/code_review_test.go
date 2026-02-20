package hooks

import "testing"

func TestCodeReviewPatternMatching(t *testing.T) {
	tests := []struct {
		prompt string
		want   bool
	}{
		// Should match — direct review requests
		{"review my code", true},
		{"review the code", true},
		{"review this code", true},
		{"review my changes", true},
		{"review the changes", true},
		{"review this PR", true},
		{"review my pull request", true},
		{"code review", true},
		{"do a code review", true},
		{"pull request review", true},
		{"PR review", true},
		{"pr review", true},
		{"Code Review please", true},

		// Should match — indirect review requests
		{"check my code", true},
		{"check this PR", true},
		{"look at my changes", true},
		{"look at this code", true},
		{"examine the code", true},
		{"examine my changes", true},

		// Should match — mixed case
		{"Review My Code", true},
		{"CODE REVIEW", true},
		{"Check My Changes", true},

		// Should match — embedded in longer prompts
		{"Can you review my code please?", true},
		{"I'd like a code review of the latest changes", true},
		{"Please look at my changes and give feedback", true},

		// Should NOT match — unrelated prompts
		{"hello", false},
		{"build the project", false},
		{"run the tests", false},
		{"review the documentation structure", false},
		{"fix the bug", false},
		{"add a new feature", false},
		{"refactor the code", false},
		{"", false},

		// Should NOT match — partial/ambiguous
		{"review", false},
		{"check the logs", false},
		{"look at the database", false},
		{"examine the architecture", false},
	}
	for _, tt := range tests {
		t.Run(tt.prompt, func(t *testing.T) {
			got := codeReviewPattern.MatchString(tt.prompt)
			if got != tt.want {
				t.Errorf("codeReviewPattern.MatchString(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestCodeReviewHookRegistered(t *testing.T) {
	_, ok := registry["code-review"]
	if !ok {
		t.Error("code-review hook not registered")
	}
}
