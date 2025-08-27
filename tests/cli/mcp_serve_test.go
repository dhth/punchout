package cli

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMCPServeCmd(t *testing.T) {
	fx, err := newFixture()
	require.NoErrorf(t, err, "error setting up fixture: %s", err)

	defer func() {
		err := fx.cleanup()
		require.NoErrorf(t, err, "error cleaning up fixture: %s", err)
	}()

	// SUCCESSES
	t.Run("help flag works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--help",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)

		result = pathRegex.ReplaceAllString(result, `default "[PATH]"`)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("listing config works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("changing transport works", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--config-file-path", "config/good.toml",
			"--db-path", "db.db",
			"--transport", "http",
			"--http-port", "3000",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	// FAILURES
	t.Run("fails if invalid transport provided", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--transport", "blah",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})

	t.Run("fails if invalid http port provided", func(t *testing.T) {
		// GIVEN
		args := []string{
			"mcp",
			"serve",
			"--transport", "http",
			"--http-port", "blah",
			"--list-config",
		}

		// WHEN
		result, err := fx.runCmd(args)

		// THEN
		require.NoError(t, err)
		snaps.MatchStandaloneSnapshot(t, result)
	})
}
