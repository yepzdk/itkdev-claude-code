package hooks

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/jesperpedersen/picky-claude/internal/notify"
)

func init() {
	Register("notify", notifyHook)
}

// notifyHook sends a desktop notification based on the hook event.
// It is non-blocking and never prevents tool execution.
func notifyHook(input *Input) error {
	switch input.HookEventName {
	case "Stop":
		detail := lastAssistantText(input.TranscriptPath)
		if detail == "" {
			detail = "Session ended"
		}
		// Fire and forget — notification failures should not block.
		notify.Send(notify.EventSessionComplete, detail)
	}
	return nil
}

// maxNotifyLen is the maximum character length for a notification body.
const maxNotifyLen = 200

// lastAssistantText reads the transcript JSONL and returns the text content
// from the last assistant message, truncated to maxNotifyLen.
func lastAssistantText(path string) string {
	if path == "" {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var lastText string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var entry struct {
			Type    string `json:"type"`
			Message struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		if entry.Type != "assistant" || entry.Message.Role != "assistant" {
			continue
		}
		if text := extractText(entry.Message.Content); text != "" {
			lastText = text
		}
	}
	return truncate(lastText, maxNotifyLen)
}

// extractText pulls text blocks out of an assistant message content array.
func extractText(raw json.RawMessage) string {
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, " ")
}

// truncate shortens s to maxLen, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-1]) + "…"
}
