package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/itk-dev/itkdev-claude-code/internal/config"
	"github.com/itk-dev/itkdev-claude-code/internal/console"
	"github.com/itk-dev/itkdev-claude-code/internal/installer"
	"github.com/itk-dev/itkdev-claude-code/internal/installer/steps"
	"github.com/itk-dev/itkdev-claude-code/internal/session"
	"github.com/spf13/cobra"
)

var issueFlag string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch Claude Code with hooks and Endless Mode",
	Long: `Starts the console server, generates a session ID, and launches
Claude Code with the appropriate environment variables and hooks.
Signals are forwarded to Claude Code. The console server runs as a
background goroutine for the lifetime of the session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: cfg.LogLevel,
		}))

		// Auto-install if project is not set up
		if err := autoInstallIfNeeded(logger); err != nil {
			return fmt.Errorf("auto-install failed: %w", err)
		}

		// Check for required plugins
		checkRequiredPlugins(logger)

		// Find Claude Code
		claudePath, err := session.FindClaudeCode()
		if err != nil {
			return fmt.Errorf("claude code not found: %w", err)
		}

		// Generate session ID
		sessionID := session.NewID()
		sessionDir := config.SessionDir(sessionID)
		if err := session.EnsureSessionDir(sessionDir); err != nil {
			return fmt.Errorf("create session dir: %w", err)
		}

		logger.Debug("starting session", "id", sessionID)

		// Start console server as goroutine
		srv, err := console.New(cfg.Port, logger)
		if err != nil {
			return fmt.Errorf("create console server: %w", err)
		}

		readyCh := make(chan struct{})
		srvErr := make(chan error, 1)
		go func() {
			srvErr <- srv.StartWithReady(readyCh)
		}()

		// Wait for server to be ready or fail
		select {
		case <-readyCh:
			// Server is listening
		case err := <-srvErr:
			return fmt.Errorf("console server failed: %w", err)
		}

		// Get actual port (may differ from configured if port was busy)
		actualPort := srv.Port()

		// Update config files with actual port so Claude Code sees the right URL
		updatePortInConfigs(actualPort, logger)

		// Write PID file for session tracking
		session.WritePIDFile(sessionDir)
		defer session.RemovePIDFile(sessionDir)

		// Register session with console
		client := session.DefaultConsoleClient(actualPort)
		if resp, err := client.Post("/api/sessions", map[string]string{
			"id":      sessionID,
			"project": detectProject(),
		}); err == nil && resp != nil {
			resp.Body.Close()
		}

		// Build environment for Claude Code
		env := session.BuildEnv(sessionID, actualPort, issueFlag)

		// Launch Claude Code
		claudeArgs := session.BuildClaudeArgs()
		claudeArgs = append(claudeArgs, args...)

		claudeCmd := exec.Command(claudePath, claudeArgs...)
		claudeCmd.Env = env
		claudeCmd.Stdin = os.Stdin
		claudeCmd.Stdout = os.Stdout
		claudeCmd.Stderr = os.Stderr

		if err := claudeCmd.Start(); err != nil {
			srv.Stop()
			return fmt.Errorf("start claude code: %w", err)
		}

		// Forward signals to Claude Code
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			for sig := range sigCh {
				claudeCmd.Process.Signal(sig)
			}
		}()

		// Wait for Claude Code to exit
		exitErr := claudeCmd.Wait()

		// End session in console
		if resp, err := client.Post(fmt.Sprintf("/api/sessions/%s/end", sessionID), nil); err == nil && resp != nil {
			resp.Body.Close()
		}

		// Stop console server
		srv.Stop()

		signal.Stop(sigCh)
		close(sigCh)

		if exitErr != nil {
			if exitCode, ok := exitErr.(*exec.ExitError); ok {
				os.Exit(exitCode.ExitCode())
			}
			return exitErr
		}

		return nil
	},
}

// detectProject tries to determine the project name from the current directory.
func detectProject() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}

// updatePortInConfigs updates .claude/settings.json and .claude/.mcp.json
// with the actual server port. This ensures the announcement and MCP URL
// reflect the port the console is actually listening on.
func updatePortInConfigs(port int, logger *slog.Logger) {
	claudeDir := filepath.Join(".claude")

	updateSettingsAnnouncement(filepath.Join(claudeDir, "settings.json"), port, logger)
	updateMCPPort(filepath.Join(claudeDir, ".mcp.json"), port, logger)
}

// updateSettingsAnnouncement rewrites the companyAnnouncements in settings.json
// to use the given port.
func updateSettingsAnnouncement(path string, port int, logger *slog.Logger) {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Debug("skipping settings.json port update", "error", err)
		return
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		logger.Debug("skipping settings.json port update: invalid JSON", "error", err)
		return
	}

	settings["companyAnnouncements"] = []string{
		"Console: http://localhost:" + strconv.Itoa(port) + " | /spec — plan, build & verify",
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		logger.Debug("skipping settings.json port update: marshal failed", "error", err)
		return
	}

	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		logger.Debug("failed to update settings.json port", "error", err)
	}
}

// updateMCPPort rewrites the mem-search URL in .mcp.json to use the given port.
func updateMCPPort(path string, port int, logger *slog.Logger) {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Debug("skipping .mcp.json port update", "error", err)
		return
	}

	var mcpConfig map[string]any
	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		logger.Debug("skipping .mcp.json port update: invalid JSON", "error", err)
		return
	}

	servers, ok := mcpConfig["mcpServers"].(map[string]any)
	if !ok {
		return
	}

	memSearch, ok := servers["mem-search"].(map[string]any)
	if !ok {
		return
	}

	memSearch["url"] = "http://localhost:" + strconv.Itoa(port) + "/mcp"

	out, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		logger.Debug("skipping .mcp.json port update: marshal failed", "error", err)
		return
	}

	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		logger.Debug("failed to update .mcp.json port", "error", err)
	}
}

// autoInstallIfNeeded detects if the project hasn't been set up yet (missing
// .claude/settings.json) and runs the installer automatically.
func autoInstallIfNeeded(logger *slog.Logger) error {
	settingsPath := filepath.Join(".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		return nil // Already set up
	}

	fmt.Fprintln(os.Stderr, "Project not set up — running installer automatically...")
	fmt.Fprintln(os.Stderr)

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	installSteps := []installer.Step{
		&steps.Prerequisites{},
		&steps.Dependencies{},
		&steps.ShellConfig{},
		&steps.ClaudeFiles{},
		&steps.ConfigFiles{},
		&steps.Plugins{},
		&steps.VSCode{},
		&steps.Finalize{},
	}

	inst := installer.New(dir, installSteps...)
	result := inst.RunWithUI(os.Stderr)

	if !result.Success {
		return fmt.Errorf("installer failed at step %q: %w", result.FailedStep, result.Error)
	}

	logger.Debug("auto-install completed successfully")
	return nil
}

// checkRequiredPlugins warns on stderr if any required Claude Code plugins
// are not installed. This is a non-blocking check — it never prevents launch.
func checkRequiredPlugins(logger *slog.Logger) {
	missing, err := config.MissingPlugins()
	if err != nil {
		logger.Debug("could not check plugins", "error", err)
		return
	}
	for _, req := range missing {
		fmt.Fprintf(os.Stderr, "⚠ Required plugin %q is not installed.\n", req.Name)
		fmt.Fprintf(os.Stderr, "  Install it in Claude Code with:\n")
		fmt.Fprintf(os.Stderr, "    /plugin marketplace add itk-dev/itkdev-claude-plugins\n")
		fmt.Fprintf(os.Stderr, "    /plugin install %s@%s\n\n", req.Name, req.Marketplace)
	}
}

func init() {
	runCmd.Flags().StringVar(&issueFlag, "issue", "", "GitHub issue number to work on (sets ICC_ISSUE_ID)")
	rootCmd.AddCommand(runCmd)
}
