package tour

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestPages(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)

	m := newModel("/home/user/.config/punchout/punchout.toml", thm)
	m.width = minWidth
	m.height = minHeight

	for pageIndex := range m.pages {
		t.Run(fmt.Sprintf("%d", pageIndex+1), func(t *testing.T) {
			m.page = pageIndex
			snaps.MatchStandaloneSnapshot(t, ansi.Strip(m.View().Content))
		})
	}
}

func TestInsufficientDimensions(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)

	m := newModel("/home/user/.config/punchout/punchout.toml", thm)
	m.width = minWidth - 1
	m.height = minHeight - 1

	snaps.MatchStandaloneSnapshot(t, ansi.Strip(m.View().Content))
}
