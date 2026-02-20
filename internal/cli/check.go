package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/itk-dev/itkdev-claude-code/internal/session"
	"github.com/spf13/cobra"
)

// errMissingDeps is returned by the check command when required dependencies are missing.
var errMissingDeps = fmt.Errorf("required dependencies missing")

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check dependency status",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		// Create logger for preflight check (debug level for standalone command)
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))

		results := session.PreflightCheck(logger)

		if jsonOutput {
			if err := json.NewEncoder(out).Encode(results); err != nil {
				return err
			}
			// Check for missing deps even in JSON mode
			for _, r := range results {
				if r.Status == session.StatusMissing {
					return errMissingDeps
				}
			}
			return nil
		}

		// Human-readable output
		fmt.Fprintln(out, "Dependency Check:")

		// Track if any required deps are missing for exit code
		hasMissing := false

		const statusWidth = 10
		const nameWidth = 20

		for _, r := range results {
			// Map status to display label
			displayStatus := r.Status
			if r.Status == session.StatusMissing {
				displayStatus = "MISSING"
				hasMissing = true
			}

			// Pad status to fixed width
			paddedStatus := fmt.Sprintf("%-*s", statusWidth, displayStatus)

			// Calculate dots needed for alignment
			dotsNeeded := nameWidth - len(r.Name)
			if dotsNeeded < 2 {
				dotsNeeded = 2
			}
			dots := " " + strings.Repeat(".", dotsNeeded-1)

			fmt.Fprintf(out, "  %s%s%s %s\n", paddedStatus, r.Name, dots, r.Message)
		}

		if hasMissing {
			return errMissingDeps
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
