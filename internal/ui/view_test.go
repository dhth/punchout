package ui

import (
	"fmt"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	d "github.com/dhth/punchout/internal/domain"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestIssueListViewColumnRendering(t *testing.T) {
	for _, width := range []int{72, 120, 160, 200} {
		t.Run(fmt.Sprintf("at width %d", width), func(t *testing.T) {
			// GIVEN
			m := newSnapshotModel(t)

			// WHEN
			result := renderSnapshotView(&m, issueListView, width)

			// THEN
			snaps.MatchStandaloneSnapshot(t, result)
		})
	}
}

func TestUnsyncedWorklogListViewColumnRendering(t *testing.T) {
	for _, width := range []int{72, 120, 160, 200} {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			// GIVEN
			m := newSnapshotModel(t)

			// WHEN
			result := renderSnapshotView(&m, wLView, width)

			// THEN
			snaps.MatchStandaloneSnapshot(t, result)
		})
	}
}

func TestSyncedWorklogListViewColumnRendering(t *testing.T) {
	for _, width := range []int{72, 120, 160, 200} {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			// GIVEN
			m := newSnapshotModel(t)

			// WHEN
			result := renderSnapshotView(&m, syncedWLView, width)

			// THEN
			snaps.MatchStandaloneSnapshot(t, result)
		})
	}
}

func TestInsufficientDimensionsView(t *testing.T) {
	t.Run("is entered when dimensions go below threshold", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			width  int
			height int
		}{
			{name: "width", width: minWidth - 1, height: minHeight},
			{name: "height", width: minWidth, height: minHeight - 1},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m := newSnapshotModel(t)
				m.processMessage(tea.WindowSizeMsg{Width: tc.width, Height: tc.height})

				require.True(t, m.dimensionsInsufficient())
			})
		}
	})

	t.Run("renders correctly", func(t *testing.T) {
		m := newSnapshotModel(t)
		m.processMessage(tea.WindowSizeMsg{Width: minWidth - 1, Height: minHeight - 1})

		snaps.MatchStandaloneSnapshot(t, ansi.Strip(m.View().Content))
	})

	t.Run("handles window resizing", func(t *testing.T) {
		m := newSnapshotModel(t)
		require.False(t, m.dimensionsInsufficient())

		m.processMessage(tea.WindowSizeMsg{Width: minWidth - 1, Height: minHeight})
		require.True(t, m.dimensionsInsufficient())
		require.Equal(t, minWidth-1, m.terminalWidth)
		require.Equal(t, minHeight, m.terminalHeight)

		m.processMessage(tea.WindowSizeMsg{Width: minWidth, Height: minHeight - 1})
		require.True(t, m.dimensionsInsufficient())
		require.Equal(t, minWidth, m.terminalWidth)
		require.Equal(t, minHeight-1, m.terminalHeight)

		m.processMessage(tea.WindowSizeMsg{Width: minWidth, Height: minHeight})
		require.False(t, m.dimensionsInsufficient())
		require.Equal(t, minWidth, m.terminalWidth)
		require.Equal(t, minHeight, m.terminalHeight)
	})

	t.Run("allows quit keys", func(t *testing.T) {
		for _, key := range []tea.Key{
			{Text: "q", Code: 'q'},
			{Code: tea.KeyEscape},
			{Code: 'c', Mod: tea.ModCtrl},
		} {
			t.Run(key.String(), func(t *testing.T) {
				m := newSnapshotModel(t)
				m.processMessage(tea.WindowSizeMsg{Width: minWidth - 1, Height: minHeight - 1})

				cmds := m.processMessage(tea.KeyPressMsg(key))

				require.Len(t, cmds, 1)
				require.IsType(t, tea.QuitMsg{}, cmds[0]())
			})
		}
	})

	t.Run("ignores non-quit keypresses", func(t *testing.T) {
		m := newSnapshotModel(t)
		m.processMessage(tea.WindowSizeMsg{Width: minWidth - 1, Height: minHeight - 1})

		require.Empty(t, m.processMessage(tea.KeyPressMsg(tea.Key{Text: "2", Code: '2'})))

		require.Equal(t, issueListView, m.activeView)
	})

	t.Run("handles message clearing", func(t *testing.T) {
		m := newSnapshotModel(t)
		m.showHelpIndicator = true
		m.message = userMsg{id: 1, value: "clear me"}
		m.processMessage(tea.WindowSizeMsg{Width: minWidth - 1, Height: minHeight - 1})

		require.Empty(t, m.processMessage(hideHelpMsg{}))
		require.Empty(t, m.processMessage(clearUserMsgMsg{id: 1}))

		require.False(t, m.showHelpIndicator)
		require.False(t, m.message.isActive())
	})
}

func newSnapshotModel(t *testing.T) Model {
	t.Helper()

	m := newTestModel(t)

	issues := []*d.Issue{
		// Baseline with ordinary values
		{
			IssueKey:        "APP-123",
			IssueType:       "Task",
			Summary:         "Resolve intermittent application error",
			Assignee:        "Ada Lovelace",
			Status:          "In Progress",
			AggSecondsSpent: 3_600,
		},
		// Long issue type
		{
			IssueKey:        "PLATFORM-4567",
			IssueType:       "Extremely Long Custom Issue Type",
			Summary:         "Prepare platform release",
			Assignee:        "Grace Hopper",
			Status:          "Failed",
			AggSecondsSpent: 9_900,
		},
		// Assignee with non-ASCII characters
		{
			IssueKey:        "UI-89",
			IssueType:       "Ops",
			Summary:         "Update navigation layout",
			Assignee:        "後藤英一",
			Status:          "In Progress",
			AggSecondsSpent: 60,
		},
		// Issue type with non-ASCII characters
		{
			IssueKey:        "OPS-126",
			IssueType:       "機能",
			Summary:         "Improve search performance",
			Assignee:        "Alan Turing",
			Status:          "To Do",
			AggSecondsSpent: 604_800,
		},
		// Combined long and non-ASCII values
		{
			IssueKey:        "EXTRAORDINARILY-LONG-12345",
			IssueType:       "Extremely Long Custom Issue Type",
			Summary:         "Coordinate regional infrastructure rollout",
			Assignee:        "ジョン・フォン・ノイマン",
			Status:          "Waiting for External Customer Review",
			AggSecondsSpent: 86_400,
		},
	}
	issueItems := make([]list.Item, len(issues))
	for i, issue := range issues {
		issueItems[i] = issue
		m.issueMap[issue.IssueKey] = issue
	}
	m.issueList.SetItems(issueItems)
	m.issueList.Title = "▪▫▫ Issues"
	m.issueList.Styles.Title = m.styles.issueListTitle

	worklogs := []d.StoredWorklog{
		{
			ID: 1,
			Worklog: d.Worklog{
				IssueKey: "APP-123",
				BeginTS:  referenceTime.Add(-3 * time.Hour),
				EndTS:    referenceTime.Add(-2 * time.Hour),
				Comment:  "Investigated the intermittent application error",
			},
		},
		{
			ID: 2,
			Worklog: d.Worklog{
				IssueKey: "PLATFORM-4567",
				BeginTS:  referenceTime.AddDate(0, 0, -3),
				EndTS:    referenceTime.AddDate(0, 0, -2),
			},
		},
		{
			ID: 3,
			Worklog: d.Worklog{
				IssueKey: "UI-89",
				BeginTS:  referenceTime.Add(30 * time.Minute),
				EndTS:    referenceTime.Add(90 * time.Minute),
			},
		},
	}
	unsyncedItems := make([]list.Item, len(worklogs))
	syncedItems := make([]list.Item, len(worklogs))
	for i, worklog := range worklogs {
		unsyncedItems[i] = worklogListItem{StoredWorklog: worklog}
		m.unsyncedWLSecsSpent += worklog.SecsSpent()
		worklog.Synced = true
		syncedItems[i] = syncedWorklogListItem{StoredWorklog: worklog}
	}
	m.worklogList.SetItems(unsyncedItems)
	m.syncedWorklogList.SetItems(syncedItems)
	m.unsyncedWLCount = uint(len(worklogs))

	fallbackComment := "Work completed without additional details"
	m.opts.Jira.FallbackComment = &fallbackComment
	m.applyTheme(m.theme)

	return m
}

func renderSnapshotView(m *Model, view stateView, width int) string {
	m.activeView = view
	m.handleWindowResizing(tea.WindowSizeMsg{Width: width, Height: 30})

	return ansi.Strip(m.View().Content)
}
