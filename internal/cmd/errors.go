package cmd

import (
	"errors"

	"github.com/dhth/punchout/internal/mcp/tools"
)

func GetErrorDetails(err error) (string, bool) {
	var followUp string
	isUnexpected := false

	switch {
	case errors.Is(err, tools.ErrCouldntAddToolToServer):
		isUnexpected = true
	}

	return followUp, isUnexpected
}
