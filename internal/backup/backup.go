package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"

	"github.com/mbrt/backsched/internal/config"
	"github.com/mbrt/backsched/internal/exec"
)

// ComputeOutdated returns the list of outdated backups.
func ComputeOutdated(ctx context.Context, cfg config.Config, env Env) ([]Info, error) {
	var res []Info
	state := loadState(env.Sio)

	for _, bc := range cfg.Backups {
		clog := log.With().Str("backup", bc.Name).Logger()
		t, ok := state.LastBackupOf(bc.Name)
		since := env.Clock.Now().Sub(t)
		if ok && since < time.Duration(bc.Interval) {
			clog.Info().Msgf("Skipping because last backup was at: %v", t)
			continue
		}
		res = append(res, Info{
			Since:  since,
			Backup: bc,
		})
	}

	return res, nil
}

// Run runs the given backup configuration.
func Run(ctx context.Context, cfg config.Config, env Env, dry bool) error {
	backups, err := ComputeOutdated(ctx, cfg, env)
	if err != nil {
		return err
	}

	state := config.State{}
	if dry {
		// Avoid running anything.
		env.Runner = dryRunner{}
	} else {
		// Make sure we update the state at the end.
		state = loadState(env.Sio)
		defer saveState(env.Sio, state)
	}

	for _, bc := range backups {
		name := bc.Backup.Name
		clog := log.With().Str("backup", name).Logger()

		b := newExecutorFromConfig(bc.Backup, env)
		if err := b.CanExecute(ctx); err != nil {
			clog.Info().Msgf("Skipping because: %v", err)
			continue
		}
		clog.Info().Msg("Executing")
		if err := b.Run(ctx); err != nil {
			return fmt.Errorf("executing backup %q: %w", name, err)
		}
		state[name] = env.Clock.Now()
	}

	return nil
}

// Env groups together the environment a backup is ran against.
type Env struct {
	Sio    StateIOer
	Clock  clockwork.Clock
	Fs     afero.Fs
	Runner exec.Runner
}

// StateIOer abstracts away lower level save and load functionality for the
// backup state.
type StateIOer interface {
	Save(buf []byte) error
	Load() ([]byte, error)
}

// Info represents the state of a backup.
type Info struct {
	// Since contains how long ago the backup was performed.
	// Zero is a special value meaning "never".
	Since  time.Duration
	Backup config.Backup
}

func (i Info) String() string {
	if i.Since == 0 {
		return fmt.Sprintf("%s: never done before", i.Backup.Name)
	}
	return fmt.Sprintf("%s: %s ago", i.Backup.Name, fmtDuration(i.Since))
}

func fmtDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}
	if d < time.Hour {
		return d.Truncate(time.Minute).String()
	}
	if d < time.Hour*24 {
		return d.Truncate(time.Hour).String()
	}
	return fmt.Sprintf("%d days", int(d/time.Hour/24))
}

func loadState(sio StateIOer) config.State {
	buf, err := sio.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Loading state")
		return config.State{}
	}
	state, err := config.LoadState(buf)
	if err != nil {
		log.Error().Err(err).Msg("Parsing state")
		return config.State{}
	}
	return state
}

func saveState(sio StateIOer, s config.State) {
	buf, err := s.Save()
	if err != nil {
		log.Error().Err(err).Msgf("Serializing state file")
	}
	if err := sio.Save(buf); err != nil {
		log.Error().Err(err).Msg("Writing state file")
	}
}

func newExecutorFromConfig(bc config.Backup, env Env) exec.Executor {
	var reqs []exec.Requirement
	for _, r := range bc.Requires {
		if r.Path != nil {
			reqs = append(reqs, exec.DirExists{Path: *r.Path})
		}
	}
	var cmds []exec.Cmd
	for _, c := range bc.Commands {
		cmds = append(cmds, exec.Cmd{
			Exe:  c.Cmd,
			Args: c.Args,
			Env:  c.Env,
		})
	}

	return exec.Executor{
		Cfg: exec.Config{
			Reqs: reqs,
			Cmds: cmds,
		},
		Fs:     env.Fs,
		Runner: env.Runner,
	}
}

type dryRunner struct{}

// Run runs a command as a subprocess.
func (dryRunner) Run(ctx context.Context, cmd string, env map[string]string, args []string) error {
	log.Info().
		Str("cmd", cmd).
		Str("args", fmt.Sprintf("%v", args)).
		Str("env", fmt.Sprintf("%v", env)).
		Msg("Would have run a command")
	return nil
}
