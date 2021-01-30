package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Backup contains all the information about a backup.
type Backup struct {
	cfg BackupCfg
	env map[string]string
	ex  Executor
}

// NewBackup creates a new instance of Backup.
func NewBackup(cfg BackupCfg, ex Executor, secrets map[string]string) Backup {
	// Copy the env map along with the secrets.
	env := map[string]string{}
	for k, v := range cfg.Env {
		env[k] = v
	}
	for k, v := range secrets {
		env[k] = v
	}
	return Backup{
		cfg: cfg,
		env: env,
		ex:  ex,
	}
}

// CanExecute returns true if the backup satisfies all requirements.
func (b Backup) CanExecute() error {
	for _, req := range b.cfg.Requires {
		if req.Path != nil {
			if !b.ex.DirExists(*req.Path) {
				return fmt.Errorf("directory %q is not present", *req.Path)
			}
		}
	}
	return nil
}

// Run runs the backup.
func (b Backup) Run() error {
	return b.ex.RunCommand(b.cfg.Cmd, b.env, b.cfg.Args)
}

// Executor executes the backup.
type Executor interface {
	DirExists(path string) bool
	RunCommand(cmd string, env map[string]string, args []string) error
}

// OSExecutor executes the commands on the local system.
type OSExecutor struct{}

// DirExists returns true if the given path exists and it's a directory.
func (OSExecutor) DirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// RunCommand runs a command.
func (OSExecutor) RunCommand(cmd string, env map[string]string, args []string) error {
	sp := exec.Command(cmd, args...)
	sp.Env = toOSEnv(env)
	sp.Stdout = os.Stdout
	sp.Stderr = os.Stderr

	fmt.Printf("Running %s %v\n", cmd, args)
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
