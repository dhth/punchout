package cli

import "regexp"

var pathRegex = regexp.MustCompile(`default "([^"]*/[^"]*)"`)
