// Package config provides centralized configuration for the application.
//
// To rename the product, change the constants in this file, update the go.mod
// module path, and the BINARY_NAME in the Makefile. Everything else propagates
// automatically.
package config

const (
	// BinaryName is the CLI executable name.
	BinaryName = "icc"

	// DisplayName is the human-readable product name used in banners and docs.
	DisplayName = "ITKdev Claude Code"

	// EnvPrefix is prepended to environment variable names (e.g. ICC_HOME).
	EnvPrefix = "ICC"

	// ConfigDirName is the directory name under $HOME for storing data.
	// Resolved to a full path by paths.go.
	ConfigDirName = ".icc"
)
