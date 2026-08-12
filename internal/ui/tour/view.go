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
	panel      lipgloss.Style
	pagination lipgloss.Style
	footerMode lipgloss.Style
	footerKey  lipgloss.Style
	footerHelp lipgloss.Style
}

type page struct {
	title   string
	content func(styles) string
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
		body:      lipgloss.NewStyle().Foreground(foreground),
		primary:   lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent1)),
		secondary: lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Accent2)),
		success:   lipgloss.NewStyle().Foreground(lipgloss.Color(thm.Success)),
		muted:     lipgloss.NewStyle().Foreground(muted),
		code:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(thm.Accent5)),
		panel: lipgloss.NewStyle().
			Width(24).
			Height(3).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(muted),
		pagination: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(6).
			Foreground(muted).
			Background(background),
		footerMode: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(background).
			Background(lipgloss.Color(thm.Accent4)),
		footerKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(thm.Accent5)).
			Background(background),
		footerHelp: lipgloss.NewStyle().
			Foreground(muted).
			Background(background),
	}
}

func (m model) View() tea.View {
	var content string
	if m.dimensionsInsufficient() {
		content = m.renderInsufficientDimensions()
	} else {
		page := renderPage(m.pages[m.page], m.styles)
		footer := m.renderFooter()
		contentHeight := m.height - lipgloss.Height(footer)
		if contentHeight < lipgloss.Height(page) {
			contentHeight = lipgloss.Height(page)
		}
		page = lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, page)
		content = lipgloss.JoinVertical(lipgloss.Left, page, footer)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.BackgroundColor = lipgloss.Color(m.theme.Background)
	v.ForegroundColor = lipgloss.Color(m.theme.Foreground)
	return v
}

func (m model) renderInsufficientDimensions() string {
	return fmt.Sprintf(`
  Terminal size too small

  Current:  %d × %d
  Required: %d × %d

  Resize the terminal to continue.

  Press q, esc, or ctrl+c to exit.
`, m.width, m.height, minWidth, minHeight)
}

func renderPage(page page, styles styles) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.titleBadge.Render(page.title),
		"",
		page.content(styles),
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
		"",
		styles.muted.Render("Press space to continue"),
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

func renderTimeTracking(styles styles) string {
	issue := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.primary.Render("▸ PROJ-142"),
		styles.body.Render("  Improve auth flow"),
		"",
	)
	timer := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.muted.Render("tracking"),
		styles.primary.Render("PROJ-142"),
		styles.code.Render("00:42:16"),
	)
	worklog := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.body.Render("Local worklog"),
		styles.muted.Render("Review, then sync"),
		"",
	)

	issueColumn := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.muted.Render("Choose an issue"),
		"",
		styles.panel.Render(issue),
	)
	timerColumn := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.muted.Render("Track your work"),
		"",
		styles.panel.Render(timer),
	)
	connector := styles.muted.Render("    ───▶    ")
	top := lipgloss.JoinHorizontal(lipgloss.Center, issueColumn, connector, timerColumn)
	worklogColumn := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.muted.Render("│\n▼"),
		styles.panel.Render(worklog),
	)
	bottom := lipgloss.PlaceHorizontal(lipgloss.Width(top), lipgloss.Right, worklogColumn)

	keyHint := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.code.Render("s"),
		styles.muted.Render("  starts and stops tracking"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		top,
		bottom,
		"",
		keyHint,
		styles.muted.Render("Punchout keeps one active timer at a time."),
	)
}

func renderCompletion(styles styles) string {
	workflow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.primary.Render("Configure."),
		" ",
		styles.secondary.Render("Track."),
		" ",
		styles.body.Render("Review."),
		" ",
		styles.success.Render("Sync."),
	)
	finish := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.muted.Render("Press "),
		styles.code.Render("space"),
		styles.muted.Render(" to finish the tour"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.heading.Render("That's the whole loop."),
		"",
		workflow,
		"",
		finish,
	)
}

func (m model) renderFooter() string {
	var navigation string
	switch {
	case m.page == 0:
		navigation = m.renderHint("l/→/space", "next")
	case m.page == len(m.pages)-1:
		navigation = m.renderHint("h/←", "back")
	default:
		navigation = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderHint("h/←", "back"),
			m.renderHint("l/→/space", "next"),
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
		m.styles.pagination.Render(m.pagination()),
		hints,
	)

	return footerContent
}

func (m model) renderHint(key, label string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.styles.footerKey.Render(key),
		" ",
		m.styles.footerHelp.Render(label),
		"   ",
	)
}
