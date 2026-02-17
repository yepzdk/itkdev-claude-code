package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallGlobalSettings_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	if err := installGlobalSettings(path, "/usr/local/bin/picky", 41777); err != nil {
		t.Fatalf("installGlobalSettings failed: %v", err)
	}

	settings := readSettingsFile(t, path)

	// Check permissions were added
	perms := getPermissionsAllow(t, settings)
	wantPerms := []string{
		"Bash(picky *)",
		"Skill(spec)",
		"Skill(spec-plan)",
		"Skill(spec-implement)",
		"Skill(spec-verify)",
	}
	for _, want := range wantPerms {
		if !containsString(perms, want) {
			t.Errorf("permissions.allow missing %q", want)
		}
	}

	// Check statusLine
	sl, ok := settings["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("statusLine should be a map")
	}
	if sl["type"] != "command" {
		t.Errorf("statusLine.type = %v, want command", sl["type"])
	}

	// Check companyAnnouncements
	ann, ok := settings["companyAnnouncements"].([]any)
	if !ok || len(ann) == 0 {
		t.Fatal("companyAnnouncements should be a non-empty array")
	}
}

func TestInstallGlobalSettings_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Create existing settings with user entries
	existing := map[string]any{
		"customSetting": true,
		"permissions": map[string]any{
			"allow": []any{"Bash(my-tool *)"},
			"deny":  []any{},
		},
	}
	writeSettingsFile(t, path, existing)

	if err := installGlobalSettings(path, "/usr/local/bin/picky", 41777); err != nil {
		t.Fatalf("installGlobalSettings failed: %v", err)
	}

	settings := readSettingsFile(t, path)

	// User's custom setting preserved
	if settings["customSetting"] != true {
		t.Error("customSetting should be preserved")
	}

	// User's existing permission preserved
	perms := getPermissionsAllow(t, settings)
	if !containsString(perms, "Bash(my-tool *)") {
		t.Error("existing permission Bash(my-tool *) should be preserved")
	}

	// Picky permissions added
	if !containsString(perms, "Skill(spec)") {
		t.Error("Skill(spec) should be added")
	}
}

func TestInstallGlobalSettings_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	installGlobalSettings(path, "/usr/local/bin/picky", 41777)
	installGlobalSettings(path, "/usr/local/bin/picky", 41777)

	settings := readSettingsFile(t, path)
	perms := getPermissionsAllow(t, settings)

	// Count occurrences of Skill(spec) — should be exactly 1
	count := 0
	for _, p := range perms {
		if p == "Skill(spec)" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Skill(spec) should appear once, got %d", count)
	}
}

func TestUninstallGlobalSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Install first
	installGlobalSettings(path, "/usr/local/bin/picky", 41777)

	// Then uninstall
	if err := uninstallGlobalSettings(path); err != nil {
		t.Fatalf("uninstallGlobalSettings failed: %v", err)
	}

	settings := readSettingsFile(t, path)

	// Permissions should be empty
	perms := getPermissionsAllow(t, settings)
	for _, p := range perms {
		if p == "Skill(spec)" || p == "Bash(picky *)" {
			t.Errorf("picky permission %q should be removed", p)
		}
	}

	// StatusLine should be removed
	if _, ok := settings["statusLine"]; ok {
		t.Error("statusLine should be removed")
	}

	// CompanyAnnouncements should be removed
	if _, ok := settings["companyAnnouncements"]; ok {
		t.Error("companyAnnouncements should be removed")
	}
}

func TestUninstallGlobalSettings_PreservesOtherEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Create settings with both picky and user entries
	existing := map[string]any{
		"customSetting": true,
		"permissions": map[string]any{
			"allow": []any{
				"Bash(my-tool *)",
				"Bash(picky *)",
				"Skill(spec)",
			},
		},
		"statusLine": map[string]any{
			"type":    "command",
			"command": "/usr/local/bin/picky statusline",
		},
		"companyAnnouncements": []any{
			"Console: http://localhost:41777 | /spec — plan, build & verify",
		},
	}
	writeSettingsFile(t, path, existing)

	uninstallGlobalSettings(path)

	settings := readSettingsFile(t, path)

	// User's custom setting preserved
	if settings["customSetting"] != true {
		t.Error("customSetting should be preserved")
	}

	// User's permission preserved
	perms := getPermissionsAllow(t, settings)
	if !containsString(perms, "Bash(my-tool *)") {
		t.Error("Bash(my-tool *) should be preserved")
	}
}

func TestUninstallGlobalSettings_NoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Should not error on missing file
	if err := uninstallGlobalSettings(path); err != nil {
		t.Fatalf("uninstallGlobalSettings should not error on missing file: %v", err)
	}
}

// --- helpers ---

func readSettingsFile(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON in %s: %v", path, err)
	}
	return m
}

func writeSettingsFile(t *testing.T, path string, m map[string]any) {
	t.Helper()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func getPermissionsAllow(t *testing.T, settings map[string]any) []string {
	t.Helper()
	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		return nil
	}
	allow, ok := perms["allow"].([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, v := range allow {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
