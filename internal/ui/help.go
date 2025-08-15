package ui

import "fmt"

var helpText = fmt.Sprintf(`
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
	helpHeaderStyle.Render("punchout Reference Manual"),
	helpSectionStyle.Render(`
  (scroll line by line with j/k/arrow keys or by half a page with <c-d>/<c-u>)

  punchout has 5 panes:
    - Issues List View                      Shows you issues matching your JQL query
    - Worklog List View                     Shows you your worklog entries; you sync these entries
                                                to JIRA from here
    - Worklog Entry/Update View             You enter/update a worklog entry from here
    - Synced Worklog List View              You view the worklog entries synced to JIRA here
    - Help View (this one)
`),
	helpHeaderStyle.Render("Keyboard Shortcuts"),
	helpHeaderStyle.Render("General"),
	helpSectionStyle.Render(`
    1                                       Switch to Issues List View
    2                                       Switch to Worklog List View
    3                                       Switch to Synced Worklog List View
    <tab>                                   Go to next view/form entry
    <shift+tab>                             Go to previous view/form entry
    q/<ctrl+c>                              Go back/reset filtering/quit
    <esc>                                   Cancel form/quit
    ?                                       Show help view
`),
	helpHeaderStyle.Render("General List Controls"),
	helpSectionStyle.Render(`
    k/<Up>                                  Move cursor up
    j/<Down>                                Move cursor down
    h<Left>                                 Go to previous page
    l<Right>                                Go to next page
    /                                       Start filtering
`),
	helpHeaderStyle.Render("Issue List View"),
	helpSectionStyle.Render(`
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
	helpHeaderStyle.Render("Worklog List View"),
	helpSectionStyle.Render(`
    <ctrl+s>/u                              Update worklog entry
    <ctrl+d>                                Delete worklog entry
    s                                       Sync all visible entries to JIRA
    <ctrl+r>                                Refresh list
`),
	helpHeaderStyle.Render("Worklog Entry/Update View"),
	helpSectionStyle.Render(`
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
	helpHeaderStyle.Render("Synced Worklog List View"),
	helpSectionStyle.Render(`
    <ctrl+r>                                Refresh list
`),
)
