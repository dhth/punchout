package ui

import "fmt"

var (
	helpText = fmt.Sprintf(`
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
    - Worklog Entry View                    You enter/update a worklog entry from here
    - Synced Worklog Entry View             You view the worklog entries synced to JIRA
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
      ?                                     Show help view
`),
		helpHeaderStyle.Render("General List Controls"),
		helpSectionStyle.Render(`
    h/<Up>                                  Move cursor up
    k/<Down>                                Move cursor down
    h<Left>                                 Go to previous page
    l<Right>                                Go to next page
    /                                       Start filtering
`),
		helpHeaderStyle.Render("Issue List View"),
		helpSectionStyle.Render(`
    s                                       Toggle recording time on the currently selected issue,
                                                will open up a form to record a comment on the second
                                            "s" keypress
    <ctrl+s>                                Add manual worklog entry
    <ctrl+t>                                Go to currently tracked item
    <ctrl+x>                                Discard currently active recording
`),
		helpHeaderStyle.Render("Worklog List View"),
		helpSectionStyle.Render(`
    <ctrl+s>                                Update worklog entry
    <ctrl+d>                                Delete worklog entry
    s                                       Sync all visible entries to JIRA
    <ctrl+r>                                Refresh list
`),
		helpHeaderStyle.Render("Worklog Entry View"),
		helpSectionStyle.Render(`
    enter                                   Save worklog entry
    k                                       Move timestamp backwards by one minute
    j                                       Move timestamp forwards by one minute
    K                                       Move timestamp backwards by five minutes
    J                                       Move timestamp forwards by five minutes
`),
		helpHeaderStyle.Render("Synced Worklog Entry View"),
		helpSectionStyle.Render(`
    <ctrl+r>                                Refresh list
`),
	)
)
