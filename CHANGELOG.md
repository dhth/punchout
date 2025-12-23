# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.3.0] - Dec 24, 2025

### Added

- An MCP server that allows for creating worklogs and syncing them to JIRA
- Allow syncing worklog timestamps with each other in workflow create/edit form
- Show time duration being recorded in workflow create/edit form
- Allow quickly finishing a worklog (ends worklog at current time, doesn't ask
    for comment)

## [v1.2.0] - Jan 16, 2025

### Added

- Allow for quickly switching actively tracked issue
- Add support for fallback comments
- Allow updating active worklog entry
- Add support for JIRA Cloud installation
- Allow shifting timestamps for worklog entries using h/j/k/l/J/K
- Show time spent on unsynced worklog entries

### Changed

- Save UTC timestamps in the database
- Allow going back views instead of quitting directly
- Improved error handling
- Upgrade to go 1.23.4
- Dependency upgrades

## [v1.1.0] - Jul 2, 2024

### Added

- Allow tweaking time when saving worklog
- Add first time help, "tracking started since" indicator
- Show indicator for currently tracked item
- Show unsynced count
- Add more colors for issue type
- Dependency upgrades

[unreleased]: https://github.com/dhth/punchout/compare/v1.3.0...HEAD
[v1.3.0]: https://github.com/dhth/punchout/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/dhth/punchout/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/dhth/punchout/compare/v1.0.0...v1.1.0
