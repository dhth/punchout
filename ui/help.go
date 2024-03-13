package ui

var (
	HelpText = `
TUI Reference Manual

punchout has 2 sections:
- Issues List View
- Worklog List View

Keyboard Shortcuts:

General
    1          Switch to Issues List View
    2          Switch to Worklog List View

General List Controls
    h/<Up>      Move cursor up
    k/<Down>    Move cursor down
    h<Left>     Go to previous page
    l<Right>    Go to next page
    /           Start filtering

Issue List View
    s           Toggle recording time on the currently selected issue,
                will open up a form to record a comment on the second
                "s" keypress
    <ctrl+s>    Add manual worklog entry

Worklog List View
    <ctrl+d>    Delete worklog entry
    s           Sync all visible entries to JIRA
    <ctrl+r>    Refresh list
`
)
