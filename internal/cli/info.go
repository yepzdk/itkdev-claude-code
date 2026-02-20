package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/itk-dev/itkdev-claude-code/internal/config"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show capabilities, hooks, commands, and MCP tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		info := buildInfo()

		if jsonOutput {
			return json.NewEncoder(out).Encode(info)
		}

		fmt.Fprintf(out, "%s v%s\n\n", info.Name, info.Version)

		fmt.Fprintln(out, "Features:")
		for _, f := range info.Features {
			fmt.Fprintf(out, "  %-18s %s\n", f.Name, f.Description)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Commands:")
		for _, c := range info.Commands {
			fmt.Fprintf(out, "  %-18s %s\n", c.Name, c.Description)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Hooks:")
		for _, h := range info.Hooks {
			fmt.Fprintf(out, "  %-18s %s\n", h.Name, h.Description)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "MCP Tools:")
		for _, m := range info.MCPTools {
			fmt.Fprintf(out, "  %-18s %s\n", m.Name, m.Description)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Skills:")
		for _, s := range info.Skills {
			fmt.Fprintf(out, "  %-18s %s\n", s.Name, s.Description)
		}

		fmt.Fprintln(out)
		fmt.Fprintf(out, "Console: http://localhost:%s\n", strconv.Itoa(config.DefaultPort))
		fmt.Fprintf(out, "Docs:    docs/usage.md\n")

		return nil
	},
}

// InfoData holds all capability information for JSON output.
type InfoData struct {
	Name     string      `json:"name"`
	Version  string      `json:"version"`
	Features []InfoEntry `json:"features"`
	Commands []InfoEntry `json:"commands"`
	Hooks    []InfoEntry `json:"hooks"`
	MCPTools []InfoEntry `json:"mcp_tools"`
	Skills   []InfoEntry `json:"skills"`
}

// InfoEntry is a name-description pair used across all info sections.
type InfoEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func buildInfo() InfoData {
	return InfoData{
		Name:    config.DisplayName,
		Version: config.Version(),
		Features: []InfoEntry{
			{"Quality Hooks", "Lint, format, and TDD enforcement on every edit"},
			{"Context Mgmt", "Context monitor with Endless Mode at 90%"},
			{"Memory", "Persistent observations with semantic search"},
			{"Workflows", "/spec (plan → implement → verify)"},
			{"Web Console", "Observations, sessions, plans, and search"},
			{"Worktree", "Git worktree isolation for safe parallel work"},
		},
		Commands: []InfoEntry{
			{"icc run", "Launch Claude Code with hooks and memory"},
			{"icc install", "Set up project configuration"},
			{"icc serve", "Standalone console server"},
			{"icc info", "Show this capabilities reference"},
			{"icc greet", "Print the welcome banner"},
			{"icc worktree", "Git worktree management"},
			{"icc session list", "List sessions"},
			{"icc check-context", "Show current context usage"},
			{"icc send-clear", "Send clear signal to session"},
		},
		Hooks: []InfoEntry{
			{"file-checker", "Language-aware lint and format on file writes"},
			{"tdd-enforcer", "Enforce test-first development order"},
			{"branch-guard", "Prevent direct commits to main"},
			{"context-monitor", "Track context usage, trigger Endless Mode"},
			{"tool-redirect", "Block or redirect tool calls"},
			{"spec-stop-guard", "Prevent premature stop during /spec"},
			{"spec-plan-validator", "Validate plan file structure"},
			{"spec-verify-validator", "Validate verification results"},
			{"task-tracker", "Track task creation and updates"},
			{"notify", "Desktop notifications on stop"},
		},
		MCPTools: []InfoEntry{
			{"search", "Semantic search across observations"},
			{"save_memory", "Save observations to persistent memory"},
			{"timeline", "Get chronological observation context"},
			{"get_observations", "Retrieve observations by filter"},
		},
		Skills: []InfoEntry{
			{"/spec", "Plan, implement, and verify with TDD"},
			{"/spec-plan", "Planning phase only"},
			{"/spec-implement", "Implementation phase only"},
			{"/spec-verify", "Verification phase only"},
			{"/itkdev-issue-workflow", "Autonomous GitHub issue workflow"},
		},
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
