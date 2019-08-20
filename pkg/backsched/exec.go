package backsched

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
)

// Executor executes a command.
type Executor interface {
	Input(prompt string) (string, error)
	Exec(cmd string, args ...string) error
	ExecOptions(opts ExecOptions, cmd string, args ...string) error
	Mkdir(path string, perm os.FileMode) error
	DirExists(path string) bool
}

// ExecOptions provides options to a command to execute
type ExecOptions struct {
	WorkDir string
	Env     []string
}

type osExecutor struct{}

func (c osExecutor) Input(prompt string) (string, error) {
	fmt.Print(fmt.Sprintf("%s: ", prompt))
	p, err := gopass.GetPasswd()
	if err != nil {
		return "", err
	}
	return string(p), nil
}

func (c osExecutor) Exec(cmd string, args ...string) error {
	opts := ExecOptions{"", nil}
	return c.ExecOptions(opts, cmd, args...)
}

func (c osExecutor) ExecOptions(opts ExecOptions, cmd string, args ...string) error {
	sp := exec.Command(cmd, args...)
	sp.Dir = opts.WorkDir
	sp.Env = opts.Env
	sp.Stdout = os.Stdout
	sp.Stderr = os.Stderr

	if err := sp.Start(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot start %s", cmd))
	}
	if err := sp.Wait(); err != nil {
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
