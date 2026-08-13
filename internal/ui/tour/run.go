package tour

import (
	tea "charm.land/bubbletea/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

func Run(defaultCfgPath string, thm theme.Theme) error {
	p := tea.NewProgram(newModel(defaultCfgPath, thm))
	_, err := p.Run()

	return err
}
