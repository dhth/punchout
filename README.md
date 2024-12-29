# punchout

✨ Overview
---

`punchout` takes the suck out of logging time on JIRA.

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout.gif" alt="Usage" />
</p>

💾 Install
---

**homebrew**:

```sh
brew install dhth/tap/punchout
```

**go**:

```sh
go install github.com/dhth/punchout@latest
```

⚡️ Usage
---

`punchout` can receive its configuration via command line flags, or a config
file.

### Using a config file

Create a toml file that looks like the following. The default location for this
file is `~/.config/punchout/punchout.toml`.

```toml
[jira]
jira_url = "https://jira.company.com"

# for on-premise installations
# you can use a JIRA PAT token here:
# jira_token = "XXX"

# or if you use cloud instance
# use your jira username and API token:
# jira_cloud_username = "example@example.com"
# jira_cloud_token = "XXX"

# put whatever JQL you want to query for
jql = "assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC"

# I don't know how many people will find use for this.
# I need this, since the JIRA server I use runs 5 hours behind
# the actual time, for whatever reason 🤷
# jira_time_delta_mins = 300
```

### Using command line flags

Use `punchout -h` for help.

```bash
punchout \
    [ -db-path='/path/to/punchout/db/file.db' ] \
    [ -jira-url='https://jira.company.com' ] \
    [ -jira-token='XXX' | { jira-cloud-token='XXX' jira-cloud-username='example@example.com' } ] \
    [ -jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC' ] \
    [ -jira-time-delta-mins='300' ] \
    [ -config-file-path='/path/to/punchout/config/file.toml' ] \
    [ -list-config ]
```

Both the config file and the command line flags can be used in conjunction, but
the latter will take precedence over the former.

```bash
punchout \
    -config-file-path='/path/to/punchout/config/file.toml' \
    -jira-token='XXX' \
    -jql='assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC'
```

🖥️ Screenshots
---

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-1.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-2.png" alt="Usage" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/punchout-3.png" alt="Usage" />
</p>

📋 Reference Manual
---

```
punchout Reference Manual

punchout has 5 panes:
  - Issues List View                      Shows you issues matching your JQL query
  - Worklog List View                     Shows you your worklog entries; you sync these entries
                                          to JIRA from here
  - Worklog Entry View                    You enter/update a worklog entry from here
  - Synced Worklog Entry View             You view the worklog entries synced to JIRA

  - Help View (this one)

Keyboard Shortcuts

General

  1                                       Switch to Issues List View
  2                                       Switch to Worklog List View
  3                                       Switch to Synced Worklog List View
  <tab>                                   Go to next view/form entry
  <shift+tab>                             Go to previous view/form entry
    ?                                     Show help view

General List Controls

  k/<Up>                                  Move cursor up
  j/<Down>                                Move cursor down
  h<Left>                                 Go to previous page
  l<Right>                                Go to next page
  /                                       Start filtering

Issue List View

  s                                       Toggle recording time on the currently selected issue,
                                              will open up a form to record a comment on the second
                                          "s" keypress
  <ctrl+s>                                Add manual worklog entry
  <ctrl+t>                                Go to currently tracked item
  <ctrl+x>                                Discard currently active recording

Worklog List View

  <ctrl+s>                                Update worklog entry
  <ctrl+d>                                Delete worklog entry
  s                                       Sync all visible entries to JIRA
  <ctrl+r>                                Refresh list

Worklog Entry View

  enter                                   Save worklog entry
  k                                       Move timestamp backwards by one minute
  j                                       Move timestamp forwards by one minute
  K                                       Move timestamp backwards by five minutes
  J                                       Move timestamp forwards by five minutes

Synced Worklog Entry View

  <ctrl+r>                                Refresh list

```

Acknowledgements
---

`punchout` is built using the awesome TUI framework [bubbletea][1].

[1]: https://github.com/charmbracelet/bubbletea
[2]: https://community.atlassian.com/t5/Atlassian-Migration-Program/Product-features-comparison-Atlassian-Cloud-vs-on-premise/ba-p/1918147
