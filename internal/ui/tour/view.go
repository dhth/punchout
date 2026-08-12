package tour

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

type styles struct {
	titleBadge lipgloss.Style
	heading    lipgloss.Style
	body       lipgloss.Style
	primary    lipgloss.Style
	secondary  lipgloss.Style
	success    lipgloss.Style
	muted      lipgloss.Style
	code       lipgloss.Style
	pagination lipgloss.Style
	footerMode lipgloss.Style
	footerHint lipgloss.Style
}

type page struct {
	title   string
	content func(styles) string
}

var pages = []page{
	{
		title:   "WELCOME",
		content: renderIntro,
	},
	{
		title:   "CONFIGURE PUNCHOUT",
		content: renderConfiguration,
	},
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
		heading: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(thm.Accent5)),
		body:       lipgloss.NewStyle().Foreground(foreground),
		primary:    lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent1)),
		secondary:  lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent2)),
		success:    lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Success)),
		muted:      lipgloss.NewStyle().Foreground(muted),
		code:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(thm.Accent5)),
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
	page := renderPage(pages[m.page], m.styles, m.pagination())

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

func renderPage(page page, styles styles, pagination string) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.titleBadge.Render(page.title),
		"",
		page.content(styles),
		"",
		styles.pagination.Render(pagination),
	)
}

func renderIntro(styles styles) string {
	journey := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.primary.Render("Record now."),
		" ",
		styles.secondary.Render("Review locally."),
		" ",
		styles.success.Render("Sync when ready."),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.heading.Render("punchout"),
		"",
		styles.body.Render("Track time against JIRA issues without breaking your flow."),
		"",
		journey,
	)
}

func renderConfiguration(styles styles) string {
	command := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.muted.Render("$ "),
		styles.primary.Render("punchout --list-config"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.body.Render("JIRA, issue, and TUI settings live in one TOML file."),
		"",
		styles.code.Render("~/.config/punchout/punchout.toml"),
		"",
		command,
		styles.muted.Render("Inspect the resolved configuration. Tokens stay redacted."),
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
