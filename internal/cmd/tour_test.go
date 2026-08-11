package cmd

import (
	"errors"
	"testing"

	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTourCommandBypassesParentValidation(t *testing.T) {
	errValidation := errors.New("validation should not run")
	parent := &cobra.Command{
		Use: "punchout",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return errValidation
		},
	}
	runCalled := false
	tourCmd := newTourCommand(func(thm theme.Theme) error {
		runCalled = true
		assert.Equal(t, theme.DefaultName, thm.Name)
		return nil
	})
	parent.AddCommand(tourCmd)
	parent.SetArgs([]string{"tour"})

	err := parent.Execute()

	require.NoError(t, err)
	assert.True(t, runCalled)
}
