package hooks

import "testing"

func TestIsGitCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"git status", true},
		{"git commit -m 'test'", true},
		{"git add . && git commit -m 'test'", true},
		{"ls -la", false},
		{"echo git", false},
		{"go test ./...", false},
		{"git", true},
		{"  git push", true},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := isGitCommand(tt.cmd); got != tt.want {
				t.Errorf("isGitCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestIsGitCommit(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"git commit -m 'test'", true},
		{"git commit --amend", true},
		{"git add . && git commit -m 'test'", true},
		{"git push", false},
		{"git status", false},
		{"echo 'git commit'", false},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := isGitCommit(tt.cmd); got != tt.want {
				t.Errorf("isGitCommit(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestIsGitPush(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"git push", true},
		{"git push origin feat/my-feature", true},
		{"git push -u origin feat/my-feature", true},
		{"git commit -m 'test'", false},
		{"git status", false},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := isGitPush(tt.cmd); got != tt.want {
				t.Errorf("isGitPush(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestIsPushToMain(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"git push origin main", true},
		{"git push origin master", true},
		{"git push --force origin main", true},
		{"git push -u origin feat/my-feature", false},
		{"git push origin feat/test", false},
		{"git push", false},
		{"git commit -m 'main'", false},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := isPushToMain(tt.cmd); got != tt.want {
				t.Errorf("isPushToMain(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestSplitChainedCommands(t *testing.T) {
	tests := []struct {
		cmd  string
		want int // expected number of parts
	}{
		{"git add .", 1},
		{"git add . && git commit -m 'test'", 2},
		{"git add .; git commit -m 'test'", 2},
		{"git add . && git commit -m 'test' && git push", 3},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			parts := splitChainedCommands(tt.cmd)
			if len(parts) != tt.want {
				t.Errorf("splitChainedCommands(%q) returned %d parts, want %d", tt.cmd, len(parts), tt.want)
			}
		})
	}
}
