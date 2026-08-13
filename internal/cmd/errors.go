package cmd

import (
	"errors"
	"fmt"

	"github.com/dhth/punchout/internal/config"
	"github.com/dhth/punchout/internal/mcp/tools"
)

func HandleError(err error) (string, bool) {
	var zero string
	switch {
	case errors.Is(err, config.ErrConfigFileNotFound):
		return fmt.Sprintf(`
Here's a sample config:

---
%s---

New to punchout? Run 'punchout tour' for a quick introduction.
`, config.SampleConfig), false
	case errors.Is(err, config.ErrParseConfigFile), errors.Is(err, config.ErrInvalidConfig):
		return `
Run 'punchout config show-sample' to view sample configuration.
`, false
	case errors.Is(err, tools.ErrCouldntAddToolToServer):
		return zero, true
	case errors.Is(err, tools.ErrCouldntConstructInputSchema):
		return zero, true
	case errors.Is(err, tools.ErrCouldntConstructOutputSchema):
		return zero, true
	}

	return zero, false
}
