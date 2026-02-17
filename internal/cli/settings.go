package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/spf13/cobra"
)

// pickyPermissions are the permissions picky adds to the global settings.
var pickyPermissions = []string{
	"Bash(picky *)",
	"Skill(spec)",
	"Skill(spec-plan)",
	"Skill(spec-implement)",
	"Skill(spec-verify)",
}

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage global Claude Code settings (~/.claude/settings.json)",
}

var settingsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Add Picky Claude entries to global Claude Code settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := globalSettingsPath()
		if err != nil {
			return err
		}
		binPath := resolvePickyBinary()
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := installGlobalSettings(path, binPath, cfg.Port); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Updated "+path)
		return nil
	},
}

var settingsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove Picky Claude entries from global Claude Code settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := globalSettingsPath()
		if err != nil {
			return err
		}
		if err := uninstallGlobalSettings(path); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Updated "+path)
		return nil
	},
}

func init() {
	settingsCmd.AddCommand(settingsInstallCmd)
	settingsCmd.AddCommand(settingsUninstallCmd)
	rootCmd.AddCommand(settingsCmd)
}

// globalSettingsPath returns ~/.claude/settings.json.
func globalSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// installGlobalSettings adds picky entries to the given settings file.
// Creates the file and parent directory if they don't exist.
// Preserves all existing entries.
func installGlobalSettings(path string, binPath string, port int) error {
	settings, err := readJSONFile(path)
	if err != nil {
		return err
	}

	// Merge permissions
	perms := ensureMap(settings, "permissions")
	allow := getStringSlice(perms, "allow")
	for _, p := range pickyPermissions {
		if !containsStr(allow, p) {
			allow = append(allow, p)
		}
	}
	perms["allow"] = toAnySlice(allow)
	settings["permissions"] = perms

	// Set statusLine
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": binPath + " statusline",
		"padding": 0,
	}

	// Set companyAnnouncements
	settings["companyAnnouncements"] = []any{
		"Console: http://localhost:" + strconv.Itoa(port) + " | /spec — plan, build & verify",
	}

	return writeJSONFile(path, settings)
}

// uninstallGlobalSettings removes picky entries from the given settings file.
// Preserves all non-picky entries.
func uninstallGlobalSettings(path string) error {
	settings, err := readJSONFile(path)
	if err != nil {
		return err
	}

	// If file didn't exist (empty settings), nothing to do
	if len(settings) == 0 {
		return nil
	}

	// Remove picky permissions
	if perms, ok := settings["permissions"].(map[string]any); ok {
		allow := getStringSlice(perms, "allow")
		var filtered []string
		for _, p := range allow {
			if !isPickyPermission(p) {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) > 0 {
			perms["allow"] = toAnySlice(filtered)
		} else {
			delete(perms, "allow")
		}
		settings["permissions"] = perms
	}

	// Remove statusLine if it's picky's
	if sl, ok := settings["statusLine"].(map[string]any); ok {
		if cmd, ok := sl["command"].(string); ok && strings.Contains(cmd, "picky") {
			delete(settings, "statusLine")
		}
	}

	// Remove companyAnnouncements if it's picky's
	if ann, ok := settings["companyAnnouncements"].([]any); ok {
		if len(ann) > 0 {
			if s, ok := ann[0].(string); ok && (strings.Contains(s, "picky") || strings.Contains(s, "/spec")) {
				delete(settings, "companyAnnouncements")
			}
		}
	}

	return writeJSONFile(path, settings)
}

// isPickyPermission returns true if the permission string is one managed by picky.
func isPickyPermission(p string) bool {
	for _, pp := range pickyPermissions {
		if p == pp {
			return true
		}
	}
	return false
}

// readJSONFile reads a JSON file into a map. Returns an empty map if the file
// doesn't exist.
func readJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[string]any), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// writeJSONFile writes a map as pretty-printed JSON to the given path.
// Creates parent directories as needed.
func writeJSONFile(path string, m map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir for %s: %w", path, err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ensureMap returns settings[key] as a map, creating it if absent.
func ensureMap(settings map[string]any, key string) map[string]any {
	if m, ok := settings[key].(map[string]any); ok {
		return m
	}
	m := make(map[string]any)
	settings[key] = m
	return m
}

// getStringSlice extracts a []string from a map's []any value.
func getStringSlice(m map[string]any, key string) []string {
	arr, ok := m[key].([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// toAnySlice converts []string to []any for JSON marshaling.
func toAnySlice(ss []string) []any {
	result := make([]any, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}

// containsStr checks if a string slice contains a value.
func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// resolvePickyBinary returns the full path to the running picky binary.
func resolvePickyBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return config.BinaryName
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	return resolved
}
