package ui

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/dhth/punchout/internal/issuecache"
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialModelUsesProvidedTheme(t *testing.T) {
	thm, err := theme.Get("dracula")
	require.NoError(t, err)

	m := InitialModel(nil, nil, nil, issuecache.Store{}, Options{}, thm, false)

	assert.Equal(t, thm, m.theme)
	assert.Equal(
		t,
		m.styles.issueListUnfetchedTitle.Render("title"),
		m.issueList.Styles.Title.Render("title"),
	)
}

func TestApplyThemeRefreshesThemeDependentUI(t *testing.T) {
	initialTheme, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)
	nextTheme, err := theme.Get("dracula")
	require.NoError(t, err)

	m := InitialModel(nil, nil, nil, issuecache.Store{}, Options{}, initialTheme, false)
	m.helpVP = viewport.New(viewport.WithWidth(120), viewport.WithHeight(20))
	m.helpVPReady = true

	oldTitle := m.styles.issueListUnfetchedTitle.Render("title")
	oldHelp := renderHelp(m.styles)

	m.applyTheme(nextTheme)

	assert.Equal(t, nextTheme, m.theme)
	assert.NotEqual(t, oldTitle, m.styles.issueListUnfetchedTitle.Render("title"))
	assert.NotEqual(t, oldHelp, renderHelp(m.styles))
	assert.Equal(
		t,
		m.styles.issueListUnfetchedTitle.Render("title"),
		m.issueList.Styles.Title.Render("title"),
	)
	assert.Equal(
		t,
		m.styles.worklogListTitle.Render("title"),
		m.worklogList.Styles.Title.Render("title"),
	)
	assert.Equal(
		t,
		m.styles.syncedWorklogListTitle.Render("title"),
		m.syncedWorklogList.Styles.Title.Render("title"),
	)
}

func TestCategoricalColorIsStableAndNamespaced(t *testing.T) {
	thm, err := theme.Get(theme.DefaultName)
	require.NoError(t, err)

	first := categoricalColor("issue-type", "Task", thm.CategoricalColors)
	second := categoricalColor("issue-type", "Task", thm.CategoricalColors)
	assignee := categoricalColor("assignee", "Task", thm.CategoricalColors)

	assert.Equal(t, fmt.Sprint(first), fmt.Sprint(second))
	assert.NotEqual(t, fmt.Sprint(first), fmt.Sprint(assignee))
}
