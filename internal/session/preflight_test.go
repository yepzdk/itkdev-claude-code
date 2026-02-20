package session

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testLogger returns a logger that discards output to keep test output clean.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPreflightCheck(t *testing.T) {
	// Create a temporary directory to use as PATH
	tmpDir := t.TempDir()

	// Create dummy git and claude binaries for testing
	for _, name := range []string{"git", "claude"} {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Set PATH to only include our temp directory
	t.Setenv("PATH", tmpDir)

	results := PreflightCheck(testLogger())

	// Verify results structure
	if len(results) == 0 {
		t.Fatal("expected results, got empty slice")
	}

	// Check for git and claude (should be "ok")
	for _, name := range []string{"git", "claude"} {
		found := false
		for _, r := range results {
			if r.Name == name {
				found = true
				if r.Status != StatusOK {
					t.Errorf("%s status = %q, want %q", name, r.Status, StatusOK)
				}
				if r.Message != "found" {
					t.Errorf("%s message = %q, want %q", name, r.Message, "found")
				}
			}
		}
		if !found {
			t.Errorf("%s not found in results", name)
		}
	}
}

func TestPreflightCheck_MissingRequired(t *testing.T) {
	// Set PATH to empty directory so binaries are not found
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)

	results := PreflightCheck(testLogger())

	// Check that git and claude have "missing" status
	for _, name := range []string{"git", "claude"} {
		found := false
		for _, r := range results {
			if r.Name == name && r.Status == StatusMissing {
				found = true
				if r.Message == "" {
					t.Errorf("%s missing message is empty", name)
				}
			}
		}
		if !found {
			t.Errorf("%s missing status not found in results", name)
		}
	}
}

func TestPreflightCheck_OptionalWarning(t *testing.T) {
	// Create temp dir with only required binaries
	tmpDir := t.TempDir()
	for _, name := range []string{"git", "claude"} {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("PATH", tmpDir)

	results := PreflightCheck(testLogger())

	// All four optional binaries should have "warning" status
	expectedOptional := []string{"npx", "vexor", "playwright-cli", "mcp-cli"}
	for _, name := range expectedOptional {
		found := false
		for _, r := range results {
			if r.Name == name {
				found = true
				if r.Status != StatusWarning {
					t.Errorf("%s status = %q, want %q", name, r.Status, StatusWarning)
				}
			}
		}
		if !found {
			t.Errorf("optional binary %q not found in results", name)
		}
	}
}

func TestCheckResult_JSONSerialization(t *testing.T) {
	result := CheckResult{
		Name:    "test-dep",
		Status:  StatusOK,
		Message: "found",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(data)

	// Verify JSON tags produce correct key names
	for _, key := range []string{`"name"`, `"status"`, `"message"`} {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("JSON output missing key %s: %s", key, jsonStr)
		}
	}

	// Verify values are present
	if !strings.Contains(jsonStr, `"test-dep"`) {
		t.Errorf("JSON output missing name value: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"ok"`) {
		t.Errorf("JSON output missing status value: %s", jsonStr)
	}
}
