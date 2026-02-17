package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotifyHookRegistered(t *testing.T) {
	_, ok := registry["notify"]
	if !ok {
		t.Error("notify hook not registered")
	}
}

func TestNotifyHookNonStopEvent(t *testing.T) {
	input := &Input{HookEventName: "PostToolUse"}
	err := notifyHook(input)
	if err != nil {
		t.Errorf("notifyHook() returned error for non-Stop event: %v", err)
	}
}

func TestNotifyHookStopEvent(t *testing.T) {
	input := &Input{HookEventName: "Stop"}
	// This will try to call osascript/notify-send. In test env it may fail,
	// but the hook should not return an error (fire and forget).
	err := notifyHook(input)
	if err != nil {
		t.Errorf("notifyHook() returned error: %v", err)
	}
}

func TestLastAssistantText(t *testing.T) {
	transcript := `{"type":"user","message":{"role":"user","content":"hello"}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"First response"}]}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Final answer here"}]}}`

	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	os.WriteFile(path, []byte(transcript), 0644)

	got := lastAssistantText(path)
	if got != "Final answer here" {
		t.Errorf("lastAssistantText() = %q, want %q", got, "Final answer here")
	}
}

func TestLastAssistantTextEmpty(t *testing.T) {
	if got := lastAssistantText(""); got != "" {
		t.Errorf("lastAssistantText(\"\") = %q, want empty", got)
	}
}

func TestLastAssistantTextNoAssistant(t *testing.T) {
	transcript := `{"type":"user","message":{"role":"user","content":"hello"}}`
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	os.WriteFile(path, []byte(transcript), 0644)

	if got := lastAssistantText(path); got != "" {
		t.Errorf("lastAssistantText() = %q, want empty", got)
	}
}

func TestLastAssistantTextToolUseOnly(t *testing.T) {
	transcript := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}]}}`
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	os.WriteFile(path, []byte(transcript), 0644)

	if got := lastAssistantText(path); got != "" {
		t.Errorf("lastAssistantText() = %q, want empty", got)
	}
}

func TestTruncate(t *testing.T) {
	short := "hello"
	if got := truncate(short, 10); got != "hello" {
		t.Errorf("truncate(%q, 10) = %q, want %q", short, got, "hello")
	}

	long := "abcdefghij"
	if got := truncate(long, 5); got != "abcd…" {
		t.Errorf("truncate(%q, 5) = %q, want %q", long, got, "abcd…")
	}
}

func TestExtractText(t *testing.T) {
	raw := []byte(`[{"type":"text","text":"Hello"},{"type":"tool_use","id":"x"},{"type":"text","text":"world"}]`)
	got := extractText(raw)
	if got != "Hello world" {
		t.Errorf("extractText() = %q, want %q", got, "Hello world")
	}
}
