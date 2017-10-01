package backsched

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// Executor executes a command.
type Executor interface {
	Exec(cmd string, args ...string) error
	Mkdir(path string, perm os.FileMode) error
	DirExists(path string) bool
}

type osExecutor struct{}

func (c osExecutor) Exec(cmd string, args ...string) error {
	sp := exec.Command(cmd, args...)
	stdout, err := sp.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "cannot pipe stdout")
	}
	if err = sp.Start(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("errors executing %s", cmd))
	}
	if _, err = io.Copy(os.Stdout, stdout); err != nil {
		return errors.Wrap(err, "cannot copy standard output")
	}
	err = sp.Wait()
	if err != nil {
		return errors.Wrap(err, "wait failed")
	}
	return nil
}

func (c osExecutor) Mkdir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (c osExecutor) DirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
