package ui

import (
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhth/punchout/internal/ui/theme"
)

type itemDelegate struct {
	delegate list.DefaultDelegate
	theme    theme.Theme
	styles   styles
}

type displayItem struct {
	item        list.Item
	title       string
	description string
}

func (i displayItem) Title() string       { return i.title }
func (i displayItem) Description() string { return i.description }
func (i displayItem) FilterValue() string { return i.item.FilterValue() }

func newItemDelegate(thm theme.Theme, styles styles, accent string) list.ItemDelegate {
	d := list.NewDefaultDelegate()
	selectionColor := lipgloss.Color(accent)
	textColor := lipgloss.Color(thm.Foreground)

	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(textColor)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(textColor)

	d.Styles.SelectedTitle = d.Styles.
		SelectedTitle.
		Foreground(selectionColor).
		BorderLeftForeground(selectionColor)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle

	return itemDelegate{
		delegate: d,
		theme:    thm,
		styles:   styles,
	}
}

func (d itemDelegate) Height() int  { return d.delegate.Height() }
func (d itemDelegate) Spacing() int { return d.delegate.Spacing() }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.delegate.Update(msg, m)
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	title, description := renderListItem(item, d.theme, d.styles)
	d.delegate.Render(w, m, index, displayItem{
		item:        item,
		title:       title,
		description: description,
	})
}
