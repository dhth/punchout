package ui

import (
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestControlCQuitsImmediately(t *testing.T) {
	for _, tc := range []struct {
		name string
		view stateView
	}{
		{name: "issue list", view: issueListView},
		{name: "worklog list", view: wLView},
		{name: "synced worklog list", view: syncedWLView},
		{name: "edit active worklog", view: editActiveWLView},
		{name: "save active worklog", view: saveActiveWLView},
		{name: "worklog entry", view: wlEntryView},
		{name: "help", view: helpView},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// GIVEN
			m := newTestModel(t)
			m.activeView = tc.view

			// WHEN
			cmds := m.processMessage(keyPress("ctrl+c"))

			// THEN
			require.Len(t, cmds, 1)
			require.IsType(t, tea.QuitMsg{}, cmds[0]())
		})
	}
}

func TestControlCQuitsWhileFiltering(t *testing.T) {
	// GIVEN
	m := newTestModel(t)
	m.issueList.SetFilterState(list.Filtering)

	// WHEN
	cmds := m.processMessage(keyPress("ctrl+c"))

	// THEN
	require.Len(t, cmds, 1)
	require.IsType(t, tea.QuitMsg{}, cmds[0]())
}

func TestQAndEscapeNavigateIdenticallyOutsideTextInput(t *testing.T) {
	for _, key := range []string{"q", "esc"} {
		t.Run(key, func(t *testing.T) {
			t.Run("clears an applied issue filter", func(t *testing.T) {
				// GIVEN
				m := newTestModel(t)
				m.issueList.SetFilterText("APP")

				// WHEN
				cmds := m.processMessage(keyPress(key))

				// THEN
				require.Empty(t, cmds)
				require.Equal(t, issueListView, m.activeView)
				require.Equal(t, list.Unfiltered, m.issueList.FilterState())
			})

			t.Run("returns from help without clearing the issue filter", func(t *testing.T) {
				// GIVEN
				m := newTestModel(t)
				m.issueList.SetFilterText("APP")
				m.activeView = helpView
				m.lastView = issueListView

				// WHEN
				cmds := m.processMessage(keyPress(key))

				// THEN
				require.Empty(t, cmds)
				require.Equal(t, issueListView, m.activeView)
				require.Equal(t, list.FilterApplied, m.issueList.FilterState())
				require.Equal(t, "APP", m.issueList.FilterValue())
			})

			for _, tc := range []struct {
				name       string
				activeView stateView
				lastView   stateView
				expected   stateView
			}{
				{name: "worklog list", activeView: wLView, expected: issueListView},
				{name: "synced worklog list", activeView: syncedWLView, expected: wLView},
				{name: "help", activeView: helpView, lastView: syncedWLView, expected: syncedWLView},
			} {
				t.Run(tc.name, func(t *testing.T) {
					// GIVEN
					m := newTestModel(t)
					m.activeView = tc.activeView
					m.lastView = tc.lastView

					// WHEN
					cmds := m.processMessage(keyPress(key))

					// THEN
					require.Empty(t, cmds)
					require.Equal(t, tc.expected, m.activeView)
				})
			}

			t.Run("quits from the issue list", func(t *testing.T) {
				// GIVEN
				m := newTestModel(t)

				// WHEN
				cmds := m.processMessage(keyPress(key))

				// THEN
				require.Len(t, cmds, 1)
				require.IsType(t, tea.QuitMsg{}, cmds[0]())
			})
		})
	}
}

func TestQIsEnteredDuringTextInput(t *testing.T) {
	t.Run("issue filter", func(t *testing.T) {
		// GIVEN
		m := newTestModel(t)
		m.issueList.SetFilterState(list.Filtering)

		// WHEN
		m.processMessage(keyPress("q"))

		// THEN
		require.Equal(t, issueListView, m.activeView)
		require.Equal(t, list.Filtering, m.issueList.FilterState())
		require.Equal(t, "q", m.issueList.FilterValue())
	})

	t.Run("worklog form", func(t *testing.T) {
		// GIVEN
		m := newTestModel(t)
		m.activeView = wlEntryView
		for i := range m.trackingInputs {
			m.trackingInputs[i].Blur()
		}
		m.trackingInputs[entryComment].Focus()

		// WHEN
		m.processMessage(keyPress("q"))

		// THEN
		require.Equal(t, wlEntryView, m.activeView)
		require.Equal(t, "q", m.trackingInputs[entryComment].Value())
	})
}

func TestEscapeCancelsTextInput(t *testing.T) {
	t.Run("issue filter", func(t *testing.T) {
		// GIVEN
		m := newTestModel(t)
		m.issueList.SetFilterState(list.Filtering)

		// WHEN
		m.processMessage(keyPress("esc"))

		// THEN
		require.Equal(t, issueListView, m.activeView)
		require.Equal(t, list.Unfiltered, m.issueList.FilterState())
	})

	t.Run("worklog form", func(t *testing.T) {
		for _, tc := range []struct {
			name         string
			view         stateView
			worklogType  worklogSaveType
			expectedView stateView
		}{
			{name: "edit active worklog", view: editActiveWLView, expectedView: issueListView},
			{name: "save active worklog", view: saveActiveWLView, expectedView: issueListView},
			{name: "insert worklog", view: wlEntryView, worklogType: worklogInsert, expectedView: issueListView},
			{name: "update worklog", view: wlEntryView, worklogType: worklogUpdate, expectedView: wLView},
		} {
			t.Run(tc.name, func(t *testing.T) {
				// GIVEN
				m := newTestModel(t)
				m.activeView = tc.view
				m.worklogSaveType = tc.worklogType

				// WHEN
				cmds := m.processMessage(keyPress("esc"))

				// THEN
				require.Empty(t, cmds)
				require.Equal(t, tc.expectedView, m.activeView)
			})
		}
	})
}

func keyPress(key string) tea.KeyPressMsg {
	switch key {
	case "esc":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	case "ctrl+c":
		return tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl})
	default:
		return tea.KeyPressMsg(tea.Key{Text: key, Code: rune(key[0])})
	}
}
