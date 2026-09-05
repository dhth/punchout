<p align="center">
  <h1 align="center">punchout</h1>
  <p align="center">
    <a href="https://github.com/dhth/punchout/actions/workflows/main.yml"><img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/dhth/punchout/main.yml?style=flat-square"></a>
    <a href="https://github.com/dhth/punchout/actions/workflows/vulncheck.yml"><img alt="Vulnerability Check" src="https://img.shields.io/github/actions/workflow/status/dhth/punchout/vulncheck.yml?style=flat-square&label=vulncheck"></a>
    <a href="https://github.com/dhth/punchout/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/release/dhth/punchout.svg?style=flat-square"></a>
    <a href="https://github.com/dhth/punchout/releases/latest"><img alt="Commits since latest release" src="https://img.shields.io/github/commits-since/dhth/punchout/latest?style=flat-square"></a>
  </p>
</p>

`punchout` takes the suck out of logging time on JIRA.

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/v1-5-0/tui-issues.png" alt="Punchout issues list" />
</p>

💾 Installation
---

### Pre-built binaries

Download a pre-built binary from the [latest
release](https://github.com/dhth/punchout/releases/latest). See [Verifying
release artifacts](#-verifying-release-artifacts) for instructions on verifying
your download.

### Install from source

You can also install from source using the `go` toolchain:

```sh
go install github.com/dhth/punchout@latest
```

🧭 Tour
---

New to `punchout`? Run the interactive tour:

```sh
punchout tour
```

The tour introduces punchout's worklog workflow, main TUI views and controls,
MCP server, and configuration. It does not require a configuration file.

⚡️ Quick start
---

Create a configuration file if you do not already have one:

```sh
mkdir -p ~/.config/punchout
punchout config show-sample > ~/.config/punchout/punchout.toml
```

Edit the generated file with your JIRA details, then validate it and start the
TUI:

```sh
punchout config validate
punchout
```

`>_` Commands
---

| Command                       | What it does                    |
|-------------------------------|---------------------------------|
| `punchout`                    | Start the TUI                   |
| `punchout tour`               | Take the interactive tour       |
| `punchout config show-sample` | Print a sample configuration    |
| `punchout config validate`    | Validate the configuration file |
| `punchout mcp serve`          | Start the MCP server            |
| `punchout help`               | Show all commands and flags     |

Run `punchout <command> --help` for details about a particular command.

🔄 Workflow
---

`punchout` lets you add worklogs to JIRA in two steps:

1. Record one or more worklogs locally.
2. Push all unsynced worklogs to JIRA.

You can do this through either the TUI or the MCP server.

⚙️ Configuration
---

`punchout` reads configuration from `~/.config/punchout/punchout.toml` by
default. Authentication settings differ between JIRA Cloud and on-premise
installations.

```toml
# String configuration values can reference environment variables. Referenced
# variables need to be set before running punchout.

# Optional. Defaults to punchout's standard database path.
# db_path = "$SOME_ENV_VAR/punchout.db"

[jira]
# Optional. Defaults to "onpremise". Allowed values: "onpremise", "cloud".
installation_type = "onpremise"

jira_url = "https://jira.company.com"
jira_token = "$PUNCHOUT_JIRA_TOKEN"

# For cloud installations, set installation_type to "cloud", use an API token,
# and provide a username.
# jira_username = "example@example.com"

# Put whatever JQL you want to use to query issues.
jql = "assignee = currentUser() AND updatedDate >= -14d ORDER BY updatedDate DESC"

# Optional. Time difference, in minutes, between your timezone and the JIRA
# server's timezone. Defaults to 0.
# jira_time_delta_mins = 300

# Optional. Used for worklogs when you do not provide a comment.
# fallback_comment = "work"

[tui]
# Optional. Defaults to false.
# use_cache_on_startup = true

# Optional. Defaults to "gruvbox-dark-hard".
# theme = "tokyonight"

[mcp]
# Optional. Defaults to "stdio". Allowed values: "stdio", "http".
# transport = "http"

# Optional. Used when transport is "http". Defaults to 18899.
# http_port = 9999
```

Command-line flags override values from the configuration file. Use a different
file with `--config-file-path`, or inspect the resolved configuration with
`--list-config`; JIRA tokens are redacted from that output.

Successful JIRA issue fetches are saved to a local cache. Set
`use_cache_on_startup` to `true` to start with the most recently cached issues
instead of immediately querying JIRA. Press `<ctrl+r>` from the issues list to
fetch the latest issues and update the cache. If the cache is unavailable,
`punchout` falls back to querying JIRA.

🖥️ TUI
---

`punchout`'s TUI lets you log time against JIRA issues and sync worklogs to
JIRA. You can track time as you work or add worklogs manually.

[![Punchout TUI demo](https://asciinema.org/a/UqtuNiBej6zGPlpW.svg)](https://asciinema.org/a/UqtuNiBej6zGPlpW)

The TUI has 5 primary views:

- **Issues List View** — Shows you issues matching your JQL query
- **Worklog List View** — Shows you your worklog entries; you sync these entries to JIRA from here
- **Worklog Entry/Update View** — You enter/update a worklog entry from here
- **Synced Worklog List View** — You view the worklog entries synced to JIRA here
- **Help View** — Shows available keymaps (as listed below)

<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/v1-5-0/tui-worklogs.png" alt="Punchout worklog list" />
</p>
<p align="center">
  <img src="https://tools.dhruvs.space/images/punchout/v1-5-0/tui-save-worklog.png" alt="Punchout worklog entry form" />
</p>

### Keyboard Shortcuts

#### General

| Mapping       | Description                        |
|---------------|------------------------------------|
| `1`           | Switch to Issues List View         |
| `2`           | Switch to Worklog List View        |
| `3`           | Switch to Synced Worklog List View |
| `<tab>`       | Go to next view/form entry         |
| `<shift+tab>` | Go to previous view/form entry     |
| `q/<ctrl+c>`  | Go back/reset filtering/quit       |
| `<esc>`       | Cancel form/quit                   |
| `[`           | Switch to previous theme           |
| `]`           | Switch to next theme               |
| `?`           | Show help view                     |

#### General List Controls

| Mapping     | Description         |
|-------------|---------------------|
| `k/<Up>`    | Move cursor up      |
| `j/<Down>`  | Move cursor down    |
| `h/<Left>`  | Go to previous page |
| `l/<Right>` | Go to next page     |
| `/`         | Start filtering     |

#### Issue List View

| Mapping    | Description                                                                                                                                |
|------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| `s`        | Toggle recording time on the currently selected issue; opens a form to record a comment on the second `s` keypress                         |
| `S`        | Quick switch recording; saves a worklog entry without a comment for the currently active issue and starts recording time for another issue |
| `f`        | Quick finish the currently active worklog                                                                                                  |
| `<ctrl+s>` | Update active worklog entry (when tracking active), or add manual worklog entry (when not tracking)                                        |
| `<ctrl+t>` | Go to currently tracked item                                                                                                               |
| `<ctrl+x>` | Discard currently active recording                                                                                                         |
| `<ctrl+b>` | Open issue in browser                                                                                                                      |
| `<ctrl+r>` | Fetch the latest issues from JIRA                                                                                                          |

#### Worklog List View

| Mapping      | Description                      |
|--------------|----------------------------------|
| `<ctrl+s>/u` | Update worklog entry             |
| `<ctrl+d>`   | Delete worklog entry             |
| `s`          | Sync all visible entries to JIRA |
| `<ctrl+r>`   | Refresh list                     |

#### Worklog Entry/Update View

| Mapping  | Description                                                  |
|----------|--------------------------------------------------------------|
| `enter`  | Save worklog entry                                           |
| `k`      | Move timestamp backwards by one minute                       |
| `j`      | Move timestamp forwards by one minute                        |
| `K`      | Move timestamp backwards by five minutes                     |
| `J`      | Move timestamp forwards by five minutes                      |
| `h`      | Move timestamp backwards by a day                            |
| `l`      | Move timestamp forwards by a day                             |
| `ctrl+s` | Sync timestamp under cursor with the other (when applicable) |

#### Synced Worklog List View

| Mapping    | Description  |
|------------|--------------|
| `<ctrl+r>` | Refresh list |

### Themes

`punchout`'s TUI comes with several built-in themes. You can see them in action
by pressing `[` or `]`. Here is a sampling of 4 built-in themes.

| Theme              | Preview                                                                                               |
|--------------------|-------------------------------------------------------------------------------------------------------|
| `catppuccin-mocha` | ![catppuccin-mocha](https://tools.dhruvs.space/images/punchout/v1-5-0/tui-theme-catppuccin-mocha.png) |
| `monokai-classic`  | ![monokai-classic](https://tools.dhruvs.space/images/punchout/v1-5-0/tui-theme-monokai-classic.png)   |
| `rose-pine-moon`   | ![rose-pine-moon](https://tools.dhruvs.space/images/punchout/v1-5-0/tui-theme-rose-pine-moon.png)     |
| `gruvbox-light`    | ![gruvbox-light](https://tools.dhruvs.space/images/punchout/v1-5-0/tui-theme-gruvbox-light.png)       |

🔌 MCP Server
---

`punchout` comes with an MCP server which allows you to automate the process of
recording worklogs and syncing them to your JIRA server.

The server uses stdio by default:

```sh
punchout mcp serve
```

It can also use Streamable HTTP:

```sh
punchout mcp serve --transport http --http-port 18899
```

The HTTP server listens on `127.0.0.1`, exposes the MCP endpoint at `/v1`, and
provides a health check at `/health`. Transport and port can also be set in the
`[mcp]` section of the configuration file.

The server provides five tools:

| Tool                    | What it does                                         |
|-------------------------|------------------------------------------------------|
| `get_jira_issues`       | Return JIRA issues matching the configured JQL       |
| `add_worklog`           | Record a worklog for an issue in punchout's database |
| `add_multiple_worklogs` | Record multiple worklogs in punchout's database      |
| `get_unsynced_worklogs` | Get unsynced worklogs from punchout's database       |
| `sync_worklogs_to_jira` | Sync all unsynced worklogs to JIRA                   |

Here's one way the MCP server can be used:

[![Punchout MCP server demo](https://asciinema.org/a/NCwiCaqLllwOFVQ3.svg)](https://asciinema.org/a/NCwiCaqLllwOFVQ3)

🔐 Verifying release artifacts
---

Each release includes checksums for all artifacts. The checksum file is signed
using [cosign](https://docs.sigstore.dev/cosign/installation/) (version
`3.1.3`).

Replace `x.y.z` below with the release version you want to verify.

1. Get the checksum and cosign signature from the release:

    ```shell
    curl -sSLO https://github.com/dhth/punchout/releases/download/vx.y.z/punchout_x.y.z_checksums.txt
    curl -sSLO https://github.com/dhth/punchout/releases/download/vx.y.z/punchout_x.y.z_checksums.txt.sigstore.json
    ```

2. Verify the checksum file's signature:

    ```shell
    cosign verify-blob \
        --bundle punchout_x.y.z_checksums.txt.sigstore.json \
        --certificate-identity-regexp 'https://github\.com/dhth/punchout/\.github/workflows/.+' \
        --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
        punchout_x.y.z_checksums.txt
    ```

3. Download the archive for your platform and validate its checksum. For example,
   for Linux x86-64:

    ```shell
    curl -sSLO https://github.com/dhth/punchout/releases/download/vx.y.z/punchout_x.y.z_linux_amd64.tar.gz
    sha256sum --ignore-missing -c punchout_x.y.z_checksums.txt
    ```

4. Once both checks pass, extract the archive:

    ```shell
    tar -xzf punchout_x.y.z_linux_amd64.tar.gz
    ./punchout -h
    ```
