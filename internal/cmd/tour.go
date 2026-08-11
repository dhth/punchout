package cmd

import (
	"github.com/dhth/punchout/internal/ui/theme"
	"github.com/spf13/cobra"
)

func newTourCommand(run func(theme.Theme) error) *cobra.Command {
	return &cobra.Command{
		Use:   "tour",
		Short: "Take a quick tour of punchout",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			thm, err := theme.Get(theme.DefaultName)
			if err != nil {
				return err
			}

			return run(thm)
		},
	}
}
