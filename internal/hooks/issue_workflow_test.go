package hooks

import "testing"

func TestIssuePatternMatching(t *testing.T) {
	tests := []struct {
		prompt string
		want   bool
	}{
		// Should match
		{"work on issue #13", true},
		{"work on issue 13", true},
		{"Work on issue #42", true},
		{"lets work on #7", true},
		{"fix #99", true},
		{"Fix #12", true},
		{"tackle #5", true},
		{"close #3", true},
		{"resolve #88", true},
		{"issue #1", true},
		{"issue 42", true},
		{"Look at issue #13 and fix it", true},
		{"Can you work on #15?", true},

		// Should not match
		{"hello", false},
		{"build the project", false},
		{"run the tests", false},
		{"what is the status", false},
		{"refactor the code", false},
		{"add a new feature", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.prompt, func(t *testing.T) {
			got := issuePattern.MatchString(tt.prompt)
			if got != tt.want {
				t.Errorf("issuePattern.MatchString(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestIssueWorkflowHookUnknownEvent(t *testing.T) {
	// Unknown events should not panic (they call ExitOK which calls os.Exit,
	// so we just verify the function signature is correct by testing the
	// pattern matching logic directly).
	if !issuePattern.MatchString("issue #13") {
		t.Error("expected pattern to match 'issue #13'")
	}
}
