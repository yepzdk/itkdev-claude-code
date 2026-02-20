package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestInfoCommand(t *testing.T) {
	jsonOutput = false // reset global flag from other tests
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"info"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command failed: %v", err)
	}

	output := buf.String()

	sections := []string{"Features:", "Commands:", "Hooks:", "MCP Tools:", "Skills:"}
	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("output missing section %q", section)
		}
	}

	entries := []string{"file-checker", "icc run", "search", "/spec"}
	for _, entry := range entries {
		if !strings.Contains(output, entry) {
			t.Errorf("output missing entry %q", entry)
		}
	}
}

func TestInfoCommandJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"info", "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info --json command failed: %v", err)
	}

	var data InfoData
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if data.Name == "" {
		t.Error("JSON name field is empty")
	}
	if data.Version == "" {
		t.Error("JSON version field is empty")
	}
	if len(data.Features) == 0 {
		t.Error("JSON features list is empty")
	}
	if len(data.Hooks) == 0 {
		t.Error("JSON hooks list is empty")
	}
	if len(data.MCPTools) == 0 {
		t.Error("JSON mcp_tools list is empty")
	}
	if len(data.Skills) == 0 {
		t.Error("JSON skills list is empty")
	}
}
