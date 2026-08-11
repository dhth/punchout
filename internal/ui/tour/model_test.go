package tour

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNavigation(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)

	for _, keys := range []struct {
		name     string
		forward  tea.KeyPressMsg
		backward tea.KeyPressMsg
	}{
		{name: "vim keys", forward: keyPress("l"), backward: keyPress("h")},
		{name: "arrow keys", forward: tea.KeyPressMsg{Code: tea.KeyRight}, backward: tea.KeyPressMsg{Code: tea.KeyLeft}},
	} {
		t.Run("moves forwards and backwards with "+keys.name, func(t *testing.T) {
			m := newModel(thm)

			updated, cmd := m.Update(keys.forward)
			m = updated.(model)
			assert.Nil(t, cmd)
			assert.Equal(t, 1, m.page)

			updated, cmd = m.Update(keys.backward)
			m = updated.(model)
			assert.Nil(t, cmd)
			assert.Equal(t, 0, m.page)
		})
	}

	t.Run("stays on the first page when moving backwards", func(t *testing.T) {
		m := newModel(thm)

		updated, cmd := m.Update(keyPress("h"))

		assert.Nil(t, cmd)
		assert.Equal(t, 0, updated.(model).page)
	})

	t.Run("quits when moving forwards from the final page", func(t *testing.T) {
		m := newModel(thm)
		m.page = len(pages) - 1

		updated, cmd := m.Update(keyPress("l"))

		assert.Equal(t, len(pages)-1, updated.(model).page)
		assert.NotNil(t, cmd)
	})
}

func TestQuitKeys(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)

	for _, key := range []tea.KeyPressMsg{
		keyPress("q"),
		{Code: tea.KeyEscape},
		{Code: 'c', Mod: tea.ModCtrl},
	} {
		t.Run(key.String(), func(t *testing.T) {
			_, cmd := newModel(thm).Update(key)

			assert.NotNil(t, cmd)
		})
	}
}

func TestThemeCycling(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)
	expected, err := theme.NextTheme(thm.Name)
	require.NoError(t, err)
	m := newModel(thm)

	updated, cmd := m.Update(keyPress("]"))
	m = updated.(model)

	assert.Nil(t, cmd)
	assert.Equal(t, expected, m.theme)
	assert.Equal(t, m.styles.product.Render("punchout"), newStyles(expected).product.Render("punchout"))
}

func TestWindowSize(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)

	updated, cmd := newModel(thm).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := updated.(model)

	assert.Nil(t, cmd)
	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestViewShowsCurrentPageAndPagination(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)
	m := newModel(thm)

	first := m.View()
	assert.Contains(t, first.Content, "WELCOME")
	assert.Contains(t, first.Content, "1 / 2")
	assert.True(t, first.AltScreen)
	assert.Equal(t, lipgloss.Color(thm.Background), first.BackgroundColor)
	assert.Equal(t, lipgloss.Color(thm.Foreground), first.ForegroundColor)

	m.page = 1
	second := m.View()
	assert.Contains(t, second.Content, "CONFIGURE PUNCHOUT")
	assert.Contains(t, second.Content, "2 / 2")
}

func keyPress(key string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
}
