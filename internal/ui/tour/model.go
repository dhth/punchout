package tour

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
)

type model struct {
	theme  theme.Theme
	styles styles
	page   int
	width  int
	height int
}

func Run(thm theme.Theme) error {
	p := tea.NewProgram(newModel(thm))
	_, err := p.Run()
	return err
}

func newModel(thm theme.Theme) model {
	return model{
		theme:  thm,
		styles: newStyles(thm),
		width:  defaultWidth,
		height: defaultHeight,
	}
}

func (model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "left", "h":
			if m.page > 0 {
				m.page--
			}
		case "right", "l":
			if m.page == len(pages)-1 {
				return m, tea.Quit
			}
			m.page++
		case "[":
			previousTheme, err := theme.PreviousTheme(m.theme.Name)
			if err == nil {
				m.applyTheme(previousTheme)
			}
		case "]":
			nextTheme, err := theme.NextTheme(m.theme.Name)
			if err == nil {
				m.applyTheme(nextTheme)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *model) applyTheme(thm theme.Theme) {
	m.theme = thm
	m.styles = newStyles(thm)
}

func (m model) pagination() string {
	return fmt.Sprintf("%d / %d", m.page+1, len(pages))
}
