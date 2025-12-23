package tools

import "errors"

var (
	ErrCouldntConstructInputSchema  = errors.New("couldn't construct input jsonschema")
	ErrCouldntConstructOutputSchema = errors.New("couldn't construct output jsonschema")
)
