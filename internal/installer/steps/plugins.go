package steps

import (
	"fmt"

	"github.com/itk-dev/itkdev-claude-code/internal/config"
	"github.com/itk-dev/itkdev-claude-code/internal/installer"
)

// Plugins checks that required Claude Code plugins are installed and provides
// instructions for installing missing ones.
type Plugins struct{}

func (p *Plugins) Name() string { return "plugins" }

func (p *Plugins) Run(ctx *installer.Context) error {
	missing, err := config.MissingPlugins()
	if err != nil {
		ctx.Messages = append(ctx.Messages, "  ⚠ Could not read installed plugins (skipping check)")
		return nil
	}

	if len(missing) == 0 {
		for _, req := range config.RequiredPlugins() {
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("  ✓ %s plugin installed", req.Name))
		}
		return nil
	}

	for _, req := range missing {
		ctx.Messages = append(ctx.Messages,
			fmt.Sprintf("  ✗ %s plugin is NOT installed", req.Name),
			"    Install it in Claude Code with:",
			"      /plugin marketplace add itk-dev/itkdev-claude-plugins",
			fmt.Sprintf("      /plugin install %s@%s", req.Name, req.Marketplace),
		)
	}

	ctx.Messages = append(ctx.Messages,
		"",
		"  ⚠ Install the missing plugins in Claude Code, then restart icc run.",
	)
	return nil
}

func (p *Plugins) Rollback(_ *installer.Context) {
	// Nothing to undo — this step only checks.
}
