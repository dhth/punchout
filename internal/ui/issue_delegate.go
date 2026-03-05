package ui

import (
	"image/color"

	"charm.land/bubbles/v2/list"
)

func newItemDelegate(color color.Color) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.
		SelectedTitle.
		Foreground(color).
		BorderLeftForeground(color)
	d.Styles.SelectedDesc = d.Styles.
		SelectedTitle

	return d
}
