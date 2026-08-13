package cmd

import (
	"errors"

	"github.com/dhth/punchout/internal/mcp/tools"
)

func HandleError(err error) (string, bool) {
	var zero string
	switch {
	case errors.Is(err, tools.ErrCouldntAddToolToServer):
		return zero, true
	case errors.Is(err, tools.ErrCouldntConstructInputSchema):
		return zero, true
	case errors.Is(err, tools.ErrCouldntConstructOutputSchema):
		return zero, true
	}

	return zero, false
}
