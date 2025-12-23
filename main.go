package main

import (
	"fmt"
	"os"

	"github.com/dhth/punchout/internal/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())

		followUp, isUnexpected := cmd.GetErrorDetails(err)
		if len(followUp) > 0 {
			fmt.Fprintf(os.Stderr, `
%s
`, followUp)
		}

		if isUnexpected {
			fmt.Fprintf(os.Stderr, `
---
This error is unexpected. 
Let @dhth know about this via https://github.com/dhth/punchout/issues.
`)
		}
		os.Exit(1)
	}
}
