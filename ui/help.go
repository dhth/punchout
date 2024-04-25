package ui

import "fmt"

var (
	HelpText = fmt.Sprintf(`
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

  punchout has 4 panes:
    - Issues List View                  Shows you issues matching your JQL query
    - Worklog List View                 Shows you your worklog entries; you sync these entries
                                        to JIRA from here
    - Worklog Entry View                You enter/update a worklog entry from here
    - Help View (this one)
`),
		helpHeaderStyle.Render("Keyboard Shortcuts"),
		helpHeaderStyle.Render("General"),
		helpSectionStyle.Render(`
    1                                   Switch to Issues List View
    2                                   Switch to Worklog List View
    <tab>                               Go to next view/form entry
    <shift+tab>                         Go to previous view/form entry
      ?                                 Show help view
`),
		helpHeaderStyle.Render("General List Controls"),
		helpSectionStyle.Render(`
    h/<Up>                              Move cursor up
    k/<Down>                            Move cursor down
    h<Left>                             Go to previous page
    l<Right>                            Go to next page
    /                                   Start filtering
`),
		helpHeaderStyle.Render("Issue List View"),
		helpSectionStyle.Render(`
    s                                   Toggle recording time on the currently selected issue,
                                            will open up a form to record a comment on the second
                                        "s" keypress
    <ctrl+s>                            Add manual worklog entry
`),
		helpHeaderStyle.Render("Worklog List View"),
		helpSectionStyle.Render(`
    <ctrl+s>                            Update worklog entry
    <ctrl+d>                            Delete worklog entry
    s                                   Sync all visible entries to JIRA
    <ctrl+r>                            Refresh list
`),
		helpHeaderStyle.Render("Worklog Entry View"),
		helpSectionStyle.Render(`
    enter                               Save worklog entry
`),
	)
)
