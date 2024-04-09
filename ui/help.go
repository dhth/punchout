package ui

var (
	HelpText = `
punchout Reference Manual

punchout has 3 panes:
- Issues List View              Shows you issues matching your JQL query
- Worklog List View             Shows you your worklog entries; you sync these entries
                                to JIRA from here
- Worklog Entry View            You enter/update a worklog entry from here

Keyboard Shortcuts:

General
    1                           Switch to Issues List View
    2                           Switch to Worklog List View
    <tab>                       Go to next view/form entry
    <shift+tab>                 Go to previous view/form entry

General List Controls
    h/<Up>                      Move cursor up
    k/<Down>                    Move cursor down
    h<Left>                     Go to previous page
    l<Right>                    Go to next page
    /                           Start filtering

Issue List View
    s                           Toggle recording time on the currently selected issue,
                                will open up a form to record a comment on the second
                                "s" keypress
    <ctrl+s>                    Add manual worklog entry

Worklog List View
    <ctrl+s>                    Update worklog entry
    <ctrl+d>                    Delete worklog entry
    s                           Sync all visible entries to JIRA
    <ctrl+r>                    Refresh list

Worklog Entry View
    enter                       Save worklog entry
`
)
