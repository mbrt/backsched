package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

// Config is an Executor configuration.
type Config struct {
	Reqs []Requirement
	Cmds []Cmd
}

// Cmd represents a command to execute.
type Cmd struct {
	// Env is the environment variables to pass.
	Env map[string]string
	// Cmd is the executable to call.
	Cmd string
	// Args is the command line arguments.
	Args []string
	// Workdir is the working directory.
	Workdir string
}

// Requirement is a requirement to satisfy.
type Requirement interface {
	Check(ctx context.Context, fs afero.Fs) error
}

// DirExists is a requirement that is satisfied when the given path is present.
type DirExists struct {
	Path string
}

// Check returns an error if the path to check is not present.
func (d DirExists) Check(ctx context.Context, fs afero.Fs) error {
	ok, err := afero.DirExists(fs, d.Path)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("directory %q doesn't exist", d.Path)
}

// Executor executes the configured commands.
type Executor struct {
	Cfg    Config
	Fs     afero.Fs
	Runner Runner
}

// CanExecute returns true if the backup satisfies all requirements.
func (e Executor) CanExecute(ctx context.Context) error {
	for _, req := range e.Cfg.Reqs {
		if err := req.Check(ctx, e.Fs); err != nil {
			return err
		}
	}
	return nil
}

// Run runs the backup.
func (e Executor) Run(ctx context.Context) error {
	for _, c := range e.Cfg.Cmds {
		if err := e.Runner.Run(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// Runner is an abstraction over command execution.
type Runner interface {
	Run(ctx context.Context, cmd Cmd) error
}

// DefaultRunner executes the commands on the local system.
type DefaultRunner struct{}

// Run runs a command as a subprocess.
func (DefaultRunner) Run(ctx context.Context, cmd Cmd) error {
	sp := exec.CommandContext(ctx, cmd.Cmd, cmd.Args...)
	sp.Env = toOSEnv(cmd.Env)
	sp.Stdin = os.Stdin
	sp.Stdout = os.Stdout
	sp.Stderr = os.Stderr
	if cmd.Workdir != "" {
		sp.Dir = cmd.Workdir
	}

	log.Info().Msgf("Running %s %v\n", cmd.Cmd, cmd.Args)
	if err := sp.Start(); err != nil {
		return fmt.Errorf("starting %q: %v", cmd.Cmd, err)
	}
	if err := sp.Wait(); err != nil {
		return fmt.Errorf("waiting for command %q: %v", cmd.Cmd, err)
	}

	return nil
}

func toOSEnv(m map[string]string) []string {
	var res []string
	for k, v := range m {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}
