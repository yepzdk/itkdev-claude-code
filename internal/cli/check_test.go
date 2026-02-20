package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/itk-dev/itkdev-claude-code/internal/session"
)

func TestCheckCommand(t *testing.T) {
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = false })

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"check"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected header
	if !strings.Contains(output, "Dependency Check:") {
		t.Error("output missing 'Dependency Check:' header")
	}

	// Verify output contains some dependency names
	if len(output) < 20 {
		t.Error("output is too short, expected dependency results")
	}
}

func TestCheckCommandJSON(t *testing.T) {
	t.Cleanup(func() { jsonOutput = false })

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"check", "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check --json command failed: %v", err)
	}

	var results []session.CheckResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected results, got empty array")
	}

	// Verify result structure
	for _, r := range results {
		if r.Name == "" {
			t.Error("result has empty Name field")
		}
		if r.Status == "" {
			t.Error("result has empty Status field")
		}
	}
}

func TestCheckCommand_MissingDepsReturnError(t *testing.T) {
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = false })

	// Set PATH to empty dir so required deps are missing
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"check"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing required dependencies, got nil")
	}
}
