package ui

import "fmt"

func renderHelp(styles styles) string {
	return fmt.Sprintf(`
  %s
%s
  %s

  %s
%s
  %s
%s
  %s
%s
  %s
%s
  %s
%s
  %s
%s
`,
		styles.helpHeader.Render("punchout Reference Manual"),
		styles.helpSection.Render(`
  (scroll line by line with j/k/arrow keys or by half a page with <c-d>/<c-u>)

  punchout has 5 panes:
    - Issues List View                      Shows you issues matching your JQL query
    - Worklog List View                     Shows you your worklog entries; you sync these entries
                                                to JIRA from here
    - Worklog Entry/Update View             You enter/update a worklog entry from here
    - Synced Worklog List View              You view the worklog entries synced to JIRA here
    - Help View (this one)
`),
		styles.helpHeader.Render("Keyboard Shortcuts"),
		styles.helpHeader.Render("General"),
		styles.helpSection.Render(`
    1                                       Switch to Issues List View
    2                                       Switch to Worklog List View
    3                                       Switch to Synced Worklog List View
    <tab>                                   Go to next view/form entry
    <shift+tab>                             Go to previous view/form entry
    q/<ctrl+c>                              Go back/reset filtering/quit
    <esc>                                   Cancel form/quit
    [                                       Switch to previous theme
    ]                                       Switch to next theme
    ?                                       Show help view
`),
		styles.helpHeader.Render("General List Controls"),
		styles.helpSection.Render(`
    k/<Up>                                  Move cursor up
    j/<Down>                                Move cursor down
    h<Left>                                 Go to previous page
    l<Right>                                Go to next page
    /                                       Start filtering
`),
		styles.helpHeader.Render("Issue List View"),
		styles.helpSection.Render(`
    s                                       Toggle recording time on the currently selected issue,
                                                will open up a form to record a comment on the second
                                                "s" keypress
    S                                       Quick switch recording; will save a worklog entry without
                                                a comment for the currently active issue, and start
                                                recording time for another issue
    f                                       Quick finish the currently active worklog
    <ctrl+s>                                Update active worklog entry (when tracking active), or add
                                                manual worklog entry (when not tracking)
    <ctrl+t>                                Go to currently tracked item
    <ctrl+x>                                Discard currently active recording
    <ctrl+b>                                Open issue in browser
`),
		styles.helpHeader.Render("Worklog List View"),
		styles.helpSection.Render(`
    <ctrl+s>/u                              Update worklog entry
    <ctrl+d>                                Delete worklog entry
    s                                       Sync all visible entries to JIRA
    <ctrl+r>                                Refresh list
`),
		styles.helpHeader.Render("Worklog Entry/Update View"),
		styles.helpSection.Render(`
    enter                                   Save worklog entry
    k                                       Move timestamp backwards by one minute
    j                                       Move timestamp forwards by one minute
    K                                       Move timestamp backwards by five minutes
    J                                       Move timestamp forwards by five minutes
    h                                       Move timestamp backwards by a day
    l                                       Move timestamp forwards by a day
    ctrl+s                                  Sync timestamp under cursor with the other (when
                                                applicable)
`),
		styles.helpHeader.Render("Synced Worklog List View"),
		styles.helpSection.Render(`
    <ctrl+r>                                Refresh list
`),
	)
}
