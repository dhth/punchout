package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func newDelegateKeyMap() *issueListdelegateKeyMap {
	return &issueListdelegateKeyMap{}
}

func newItemDelegate(keys *issueListdelegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.
		SelectedTitle.
		Foreground(lipgloss.Color("#fe8019")).
		BorderLeftForeground(lipgloss.Color("#fe8019"))
	d.Styles.SelectedDesc = d.Styles.
		SelectedTitle.
		Copy()

	return d
}
