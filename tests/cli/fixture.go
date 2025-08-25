package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Fixture struct {
	tempDir string
	binPath string
}

func newFixture() (Fixture, error) {
	var zero Fixture
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return zero, fmt.Errorf("couldn't create temporary directory: %s", err.Error())
	}

	binPath := filepath.Join(tempDir, "punchout")
	buildArgs := []string{"build", "-o", binPath, "../.."}

	c := exec.Command("go", buildArgs...)
	err = c.Run()
	if err != nil {
		return zero, fmt.Errorf("couldn't build binary: %s", err.Error())
	}

	return Fixture{
		tempDir: tempDir,
		binPath: binPath,
	}, nil
}

func (f Fixture) cleanup() error {
	err := os.RemoveAll(f.tempDir)
	if err != nil {
		return fmt.Errorf("couldn't clean up temporary directory (%s): %s", f.tempDir, err.Error())
	}

	return nil
}

func (f Fixture) runCmd(args []string) (string, error) {
	c := exec.Command(f.binPath, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	c.Stdout = &stdoutBuf
	c.Stderr = &stderrBuf

	err := c.Run()
	exitCode := 0
	success := true

	if err != nil {
		success = false
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
	}

	output := fmt.Sprintf(`success: %t
exit_code: %d
----- stdout -----
%s
----- stderr -----
%s
`, success, exitCode, stdoutBuf.String(), stderrBuf.String())

	return output, nil
}
