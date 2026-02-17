package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/statusline"
	"github.com/spf13/cobra"
)

var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Format the status bar",
	Long: `Outputs a formatted status bar with git branch, plan progress, and
context usage. Gathers data from the filesystem and environment.
Optionally reads JSON from stdin to supplement.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var input statusline.Input

		// Read stdin if available (Claude Code may pipe session data)
		data, _ := io.ReadAll(os.Stdin)
		if len(data) > 0 {
			json.Unmarshal(data, &input) //nolint:errcheck
		}

		// Gather remaining data from filesystem
		workDir, _ := os.Getwd()
		sessionID := os.Getenv(config.EnvPrefix + "_SESSION_ID")
		if sessionID == "" {
			sessionID = "default"
		}
		sessionDir := config.SessionDir(sessionID)

		statusline.Gather(&input, workDir, sessionDir)

		output := statusline.Format(&input)
		if output != "" {
			fmt.Print(output)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statuslineCmd)
}
