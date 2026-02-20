package session

import (
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/itk-dev/itkdev-claude-code/internal/config"
)

// Status constants for CheckResult.
const (
	StatusOK      = "ok"
	StatusMissing = "missing"
	StatusWarning = "warning"
)

// CheckResult represents the result of checking a single dependency.
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // StatusOK, StatusMissing, or StatusWarning
	Message string `json:"message"`
}

// PreflightCheck validates that required and optional dependencies are available.
// It checks required binaries (git, claude), optional binaries (npx, vexor,
// playwright-cli, mcp-cli), and optional plugins via config.MissingPlugins().
//
// The function returns a slice of CheckResult — one for each dependency checked.
// The caller is responsible for formatting and outputting the results.
//
// Optional binary list mirrors the installer's npm packages in installer/steps/dependencies.go.
// Keep these lists in sync when adding or removing optional dependencies.
func PreflightCheck(logger *slog.Logger) []CheckResult {
	var results []CheckResult

	// Check required binaries
	requiredBinaries := []string{"git", "claude"}
	for _, bin := range requiredBinaries {
		logger.Debug("checking binary", "name", bin, "required", true)
		_, err := exec.LookPath(bin)
		if err != nil {
			results = append(results, CheckResult{
				Name:    bin,
				Status:  StatusMissing,
				Message: "not found in PATH",
			})
		} else {
			results = append(results, CheckResult{
				Name:    bin,
				Status:  StatusOK,
				Message: "found",
			})
		}
	}

	// Check optional binaries (npm globals and npx)
	// These match the packages in installer/steps/dependencies.go
	optionalBinaries := []string{"npx", "vexor", "playwright-cli", "mcp-cli"}
	for _, bin := range optionalBinaries {
		logger.Debug("checking binary", "name", bin, "required", false)
		_, err := exec.LookPath(bin)
		if err != nil {
			results = append(results, CheckResult{
				Name:    bin,
				Status:  StatusWarning,
				Message: "not found (optional)",
			})
		} else {
			results = append(results, CheckResult{
				Name:    bin,
				Status:  StatusOK,
				Message: "found",
			})
		}
	}

	// Check plugins
	missing, err := config.MissingPlugins()
	if err != nil {
		logger.Debug("could not check plugins", "error", err)
		// Gracefully skip plugin checking on error (matches existing checkRequiredPlugins behavior)
	} else {
		for _, req := range missing {
			results = append(results, CheckResult{
				Name:   req.Name,
				Status: StatusWarning,
				Message: fmt.Sprintf("not installed — install with: /plugin marketplace add itk-dev/itkdev-claude-plugins && /plugin install %s@%s",
					req.Name, req.Marketplace),
			})
		}
	}

	return results
}
