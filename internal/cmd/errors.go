package cmd

import (
	"errors"

	"github.com/dhth/punchout/internal/mcp/tools"
)

func IsErrUnexpected(err error) bool {
	switch {
	case errors.Is(err, tools.ErrCouldntAddToolToServer):
		return true
	case errors.Is(err, tools.ErrCouldntConstructInputSchema):
		return true
	case errors.Is(err, tools.ErrCouldntConstructOutputSchema):
		return true
	}
	return false
}
