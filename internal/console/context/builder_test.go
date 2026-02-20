package context

import (
	"testing"

	"github.com/itk-dev/itkdev-claude-code/internal/db"
)

func TestBuildEmpty(t *testing.T) {
	b := NewBuilder(4000)
	result := b.Build(nil, nil)
	if result != "" {
		t.Errorf("expected empty string for nil inputs, got %q", result)
	}
}

func TestBuildWithObservationsOnly(t *testing.T) {
	b := NewBuilder(4000)

	obs := []*db.Observation{
		{ID: 1, Type: "discovery", Title: "Found bug", Text: "Auth token expired"},
		{ID: 2, Type: "decision", Title: "Use JWT", Text: "Switched to JWT tokens"},
	}

	result := b.Build(obs, nil)

	if result == "" {
		t.Fatal("expected non-empty context")
	}
	if !containsAll(result, "Found bug", "Use JWT", "Auth token expired") {
		t.Errorf("result missing expected content: %s", result)
	}
}

func TestBuildWithSummariesOnly(t *testing.T) {
	b := NewBuilder(4000)

	summaries := []*db.Summary{
		{ID: 1, SessionID: "s1", Text: "Implemented authentication flow"},
	}

	result := b.Build(nil, summaries)

	if result == "" {
		t.Fatal("expected non-empty context")
	}
	if !containsAll(result, "Implemented authentication flow") {
		t.Errorf("result missing expected content: %s", result)
	}
}

func TestBuildWithBoth(t *testing.T) {
	b := NewBuilder(4000)

	obs := []*db.Observation{
		{ID: 1, Type: "discovery", Title: "Config loading", Text: "YAML config is loaded from ~/.icc/"},
	}
	summaries := []*db.Summary{
		{ID: 1, SessionID: "s1", Text: "Refactored config system"},
	}

	result := b.Build(obs, summaries)

	if !containsAll(result, "Config loading", "Refactored config system") {
		t.Errorf("result missing expected content: %s", result)
	}
}

func TestBuildTruncatesToTokenBudget(t *testing.T) {
	b := NewBuilder(100) // very tight budget

	var obs []*db.Observation
	for i := 0; i < 50; i++ {
		obs = append(obs, &db.Observation{
			ID:    int64(i),
			Type:  "discovery",
			Title: "Long observation title that takes up tokens",
			Text:  "This is a long text body that should consume token budget quite quickly when many are added",
		})
	}

	result := b.Build(obs, nil)

	tokens := EstimateTokens(result)
	if tokens > 120 { // allow some overhead for headers
		t.Errorf("result has ~%d tokens, expected <= 120", tokens)
	}
}

func TestBuildIncludesCapabilities(t *testing.T) {
	b := NewBuilder(4000)

	// Even with no observations or summaries, capabilities should not appear
	// (the builder returns empty when both inputs are nil)
	result := b.Build(nil, nil)
	if result != "" {
		t.Errorf("expected empty string for nil inputs, got %q", result)
	}

	// With observations, capabilities section should be appended
	obs := []*db.Observation{
		{ID: 1, Type: "discovery", Title: "Test", Text: "Test observation"},
	}
	result = b.Build(obs, nil)

	if !containsAll(result, "Active Features", "Quality Hooks", "Persistent Memory", "/spec") {
		t.Errorf("result missing capabilities section: %s", result)
	}
}

func TestBuildCapabilitiesOmittedWhenBudgetTight(t *testing.T) {
	b := NewBuilder(50) // very tight budget

	obs := []*db.Observation{
		{ID: 1, Type: "discovery", Title: "Important", Text: "Critical finding that takes up budget"},
	}
	result := b.Build(obs, nil)

	// Capabilities should be omitted when budget is too tight
	if containsAll(result, "Active Features") {
		t.Error("capabilities should be omitted when token budget is tight")
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input string
		min   int
		max   int
	}{
		{"", 0, 0},
		{"hello world", 1, 5},
		{"The quick brown fox jumps over the lazy dog", 5, 15},
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got < tt.min || got > tt.max {
			t.Errorf("EstimateTokens(%q) = %d, want [%d, %d]", tt.input, got, tt.min, tt.max)
		}
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
