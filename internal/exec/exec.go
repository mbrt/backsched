package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
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
	// Exe is the executable to call.
	Exe string
	// Args is the command line arguments.
	Args []string
}

// Requirement is a requirement to satisfy.
type Requirement interface {
	Check(ctx context.Context, o Os) error
}

// DirExists is a requirement that is satisfied when the given path is present.
type DirExists struct {
	Path string
}

// Check returns an error if the path to check is not present.
func (d DirExists) Check(ctx context.Context, o Os) error {
	if !o.DirExists(ctx, d.Path) {
		return fmt.Errorf("directory %q is not present", d.Path)
	}
	return nil
}

// Executor executes the configured commands.
type Executor struct {
	Cfg Config
	Os  Os
}

// CanExecute returns true if the backup satisfies all requirements.
func (e Executor) CanExecute(ctx context.Context) error {
	for _, req := range e.Cfg.Reqs {
		if err := req.Check(ctx, e.Os); err != nil {
			return err
		}
	}
	return nil
}

// Run runs the backup.
func (e Executor) Run(ctx context.Context) error {
	for _, c := range e.Cfg.Cmds {
		if err := e.Os.RunCommand(ctx, c.Exe, c.Env, c.Args); err != nil {
			return err
		}
	}
	return nil
}

// Os is an abstraction over the underlying Operating System.
type Os interface {
	DirExists(ctx context.Context, path string) bool
	RunCommand(ctx context.Context, cmd string, env map[string]string, args []string) error
}

// LocalOs executes the commands on the local system.
type LocalOs struct{}

// DirExists returns true if the given path exists and it's a directory.
func (LocalOs) DirExists(ctx context.Context, path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// RunCommand runs a command.
func (LocalOs) RunCommand(ctx context.Context, cmd string, env map[string]string, args []string) error {
	sp := exec.CommandContext(ctx, cmd, args...)
	sp.Env = toOSEnv(env)
	sp.Stdout = os.Stdout
	sp.Stderr = os.Stderr

	log.Info().Msgf("Running %s %v\n", cmd, args)
	if err := sp.Start(); err != nil {
		return fmt.Errorf("starting %q: %v", cmd, err)
	}
	if err := sp.Wait(); err != nil {
		return fmt.Errorf("waiting for command %q: %v", cmd, err)
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
