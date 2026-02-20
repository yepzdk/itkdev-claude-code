package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PluginInstallation describes one scope-specific installation of a plugin.
type PluginInstallation struct {
	Scope       string `json:"scope"`
	Version     string `json:"version"`
	InstallPath string `json:"installPath"`
}

// installedPluginsFile is the JSON structure of installed_plugins.json.
type installedPluginsFile struct {
	Version int                             `json:"version"`
	Plugins map[string][]PluginInstallation `json:"plugins"`
}

// RequiredPlugin describes a plugin that must be installed.
type RequiredPlugin struct {
	Name        string // e.g. "itkdev-tools"
	Marketplace string // e.g. "itkdev-marketplace"
}

// Key returns the lookup key used in installed_plugins.json (name@marketplace).
func (p RequiredPlugin) Key() string {
	return p.Name + "@" + p.Marketplace
}

// RequiredPlugins returns the list of plugins that ICC expects to be installed.
func RequiredPlugins() []RequiredPlugin {
	return []RequiredPlugin{
		{Name: "itkdev-tools", Marketplace: "itkdev-marketplace"},
	}
}

// ClaudePluginsDir returns the path to Claude Code's plugins directory.
func ClaudePluginsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "plugins")
}

// InstalledPlugins reads and parses ~/.claude/plugins/installed_plugins.json.
// Returns nil map (not error) if the file doesn't exist.
func InstalledPlugins() (map[string][]PluginInstallation, error) {
	dir := ClaudePluginsDir()
	if dir == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filepath.Join(dir, "installed_plugins.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var file installedPluginsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	return file.Plugins, nil
}

// IsPluginInstalled checks whether a specific plugin is present in the
// installed plugins registry.
func IsPluginInstalled(plugin RequiredPlugin) (bool, error) {
	plugins, err := InstalledPlugins()
	if err != nil {
		return false, err
	}
	if plugins == nil {
		return false, nil
	}

	installations, ok := plugins[plugin.Key()]
	return ok && len(installations) > 0, nil
}

// MissingPlugins returns the subset of RequiredPlugins that are not installed.
func MissingPlugins() ([]RequiredPlugin, error) {
	plugins, err := InstalledPlugins()
	if err != nil {
		return nil, err
	}

	var missing []RequiredPlugin
	for _, req := range RequiredPlugins() {
		installations, ok := plugins[req.Key()]
		if !ok || len(installations) == 0 {
			missing = append(missing, req)
		}
	}
	return missing, nil
}
