package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequiredPluginKey(t *testing.T) {
	p := RequiredPlugin{Name: "itkdev-tools", Marketplace: "itkdev-marketplace"}
	want := "itkdev-tools@itkdev-marketplace"
	if got := p.Key(); got != want {
		t.Errorf("Key() = %q, want %q", got, want)
	}
}

func TestRequiredPluginsNotEmpty(t *testing.T) {
	plugins := RequiredPlugins()
	if len(plugins) == 0 {
		t.Fatal("RequiredPlugins() returned empty list")
	}
}

func TestInstalledPluginsFileNotExist(t *testing.T) {
	// Point to a non-existent home dir so the file can't be found.
	t.Setenv("HOME", t.TempDir())

	plugins, err := InstalledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plugins != nil {
		t.Errorf("expected nil, got %v", plugins)
	}
}

func TestInstalledPluginsValidJSON(t *testing.T) {
	tmpHome := t.TempDir()
	pluginsDir := filepath.Join(tmpHome, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	data := `{
		"version": 2,
		"plugins": {
			"itkdev-tools@itkdev-marketplace": [
				{"scope": "user", "version": "0.3.3", "installPath": "/tmp/test"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpHome)

	plugins, err := InstalledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	installations := plugins["itkdev-tools@itkdev-marketplace"]
	if len(installations) != 1 {
		t.Fatalf("expected 1 installation, got %d", len(installations))
	}
	if installations[0].Version != "0.3.3" {
		t.Errorf("version = %q, want 0.3.3", installations[0].Version)
	}
}

func TestIsPluginInstalledTrue(t *testing.T) {
	tmpHome := t.TempDir()
	pluginsDir := filepath.Join(tmpHome, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	data := `{
		"version": 2,
		"plugins": {
			"itkdev-tools@itkdev-marketplace": [
				{"scope": "user", "version": "0.3.3", "installPath": "/tmp/test"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpHome)

	installed, err := IsPluginInstalled(RequiredPlugin{Name: "itkdev-tools", Marketplace: "itkdev-marketplace"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !installed {
		t.Error("expected plugin to be installed")
	}
}

func TestIsPluginInstalledFalse(t *testing.T) {
	tmpHome := t.TempDir()
	pluginsDir := filepath.Join(tmpHome, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	data := `{"version": 2, "plugins": {}}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpHome)

	installed, err := IsPluginInstalled(RequiredPlugin{Name: "itkdev-tools", Marketplace: "itkdev-marketplace"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if installed {
		t.Error("expected plugin to NOT be installed")
	}
}

func TestMissingPluginsAllPresent(t *testing.T) {
	tmpHome := t.TempDir()
	pluginsDir := filepath.Join(tmpHome, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	data := `{
		"version": 2,
		"plugins": {
			"itkdev-tools@itkdev-marketplace": [
				{"scope": "user", "version": "0.3.3", "installPath": "/tmp/test"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpHome)

	missing, err := MissingPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing plugins, got %v", missing)
	}
}

func TestMissingPluginsNoneInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	pluginsDir := filepath.Join(tmpHome, ".claude", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	data := `{"version": 2, "plugins": {}}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpHome)

	missing, err := MissingPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) != len(RequiredPlugins()) {
		t.Errorf("expected %d missing plugins, got %d", len(RequiredPlugins()), len(missing))
	}
}
