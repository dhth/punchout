package tour

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

const (
	minWidth  = 80
	minHeight = 24
)

type model struct {
	defaultCfgPath string
	theme          theme.Theme
	styles         styles
	pages          []page
	page           int
	width          int
	height         int
}

func (model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	}

	if m.dimensionsInsufficient() {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "left", "h":
			if m.page > 0 {
				m.page--
			}
		case "right", "l":
			if m.page < len(m.pages)-1 {
				m.page++
			}
		case "space":
			if m.page == len(m.pages)-1 {
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
	}

	return m, nil
}

func (m model) dimensionsInsufficient() bool {
	return m.width < minWidth || m.height < minHeight
}

func (m *model) applyTheme(thm theme.Theme) {
	m.theme = thm
	m.styles = newStyles(thm)
}

func (m model) pagination() string {
	return fmt.Sprintf("%d / %d", m.page+1, len(m.pages))
}

func newModel(defaultCfgPath string, thm theme.Theme) model {
	return model{
		defaultCfgPath: defaultCfgPath,
		theme:          thm,
		styles:         newStyles(thm),
		pages: []page{
			{
				title:      "Welcome",
				renderBody: renderIntro,
			},
			{
				title:      "Workflow",
				renderBody: renderWorkflow,
			},
			{
				title:      "TUI overview",
				renderBody: renderTUIOverview,
			},
			{
				title:      "MCP server",
				renderBody: renderMCPServer,
			},
			{
				title:      "Configuration",
				renderBody: renderConfiguration(defaultCfgPath),
			},
			{
				title:      "That's it",
				renderBody: renderCompletion,
			},
		},
		width:  minWidth,
		height: minHeight,
	}
}
