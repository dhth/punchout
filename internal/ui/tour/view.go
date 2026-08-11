package tour

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

type styles struct {
	titleBadge lipgloss.Style
	product    lipgloss.Style
	lead       lipgloss.Style
	record     lipgloss.Style
	review     lipgloss.Style
	sync       lipgloss.Style
	configPath lipgloss.Style
	command    lipgloss.Style
	muted      lipgloss.Style
	pagination lipgloss.Style
	footerMode lipgloss.Style
	footerHint lipgloss.Style
}

var pages = []func(model) string{
	renderIntroPage,
	renderConfigPage,
}

func newStyles(thm theme.Theme) styles {
	background := lipgloss.Color(thm.Background)
	foreground := lipgloss.Color(thm.Foreground)
	muted := lipgloss.Color(thm.Muted)

	return styles{
		titleBadge: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(background).
			Background(lipgloss.Color(thm.Accent1)),
		product: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(thm.Accent5)),
		lead:       lipgloss.NewStyle().Foreground(foreground),
		record:     lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent1)),
		review:     lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent2)),
		sync:       lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Success)),
		configPath: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(thm.Accent5)),
		command:    lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent1)),
		muted:      lipgloss.NewStyle().Foreground(muted),
		pagination: lipgloss.NewStyle().Foreground(muted),
		footerMode: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(background).
			Background(lipgloss.Color(thm.Accent4)),
		footerHint: lipgloss.NewStyle().
			Foreground(foreground).
			Background(background),
	}
}

func (m model) View() tea.View {
	page := pages[m.page](m)
	page = lipgloss.JoinVertical(
		lipgloss.Center,
		page,
		"",
		m.styles.pagination.Render(m.pagination()),
	)

	footer := m.renderFooter()
	contentHeight := m.height - lipgloss.Height(footer)
	if contentHeight < lipgloss.Height(page) {
		contentHeight = lipgloss.Height(page)
	}
	content := lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, page)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, footer))
	v.AltScreen = true
	v.BackgroundColor = lipgloss.Color(m.theme.Background)
	v.ForegroundColor = lipgloss.Color(m.theme.Foreground)
	return v
}

func renderIntroPage(m model) string {
	journey := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.styles.record.Render("Record now."),
		" ",
		m.styles.review.Render("Review locally."),
		" ",
		m.styles.sync.Render("Sync when ready."),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.styles.titleBadge.Render("WELCOME"),
		"",
		m.styles.product.Render("punchout"),
		"",
		m.styles.lead.Render("Track time against JIRA issues without breaking your flow."),
		"",
		journey,
	)
}

func renderConfigPage(m model) string {
	command := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.styles.muted.Render("$ "),
		m.styles.command.Render("punchout --list-config"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.styles.titleBadge.Render("CONFIGURE PUNCHOUT"),
		"",
		m.styles.lead.Render("JIRA, issue, and TUI settings live in one TOML file."),
		"",
		m.styles.configPath.Render("~/.config/punchout/punchout.toml"),
		"",
		command,
		m.styles.muted.Render("Inspect the resolved configuration. Tokens stay redacted."),
	)
}

func (m model) renderFooter() string {
	var navigation string
	if m.page == 0 {
		navigation = m.renderHint("l/→", "next")
	} else {
		navigation = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderHint("h/←", "back"),
			m.renderHint("l/→", "finish"),
		)
	}

	hints := lipgloss.JoinHorizontal(
		lipgloss.Top,
		navigation,
		m.renderHint("esc/q", "quit"),
		m.renderHint("[/]", "theme"),
	)
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.styles.footerMode.Render("punchout"),
		hints,
	)

	return footerContent
}

func (m model) renderHint(key, label string) string {
	return m.styles.footerHint.Render(fmt.Sprintf("  %s %s  ", key, label))
}
