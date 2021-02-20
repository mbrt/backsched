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

// Run runs the given backup configuration.
func Run(ctx context.Context, cfg config.Config, env Env) error {
	state := loadState(env.Sio)
	// Make sure we save the state at the end.
	defer saveState(env.Sio, state)

	for _, bc := range cfg.Backups {
		clog := log.With().Str("backup", bc.Name).Logger()
		if t, ok := state.LastBackupOf(bc.Name); ok && env.Clock.Now().Sub(t) < time.Duration(bc.Interval) {
			clog.Info().Msgf("Skipping because last backup was at: %v", t)
			continue
		}

		b := newExecutorFromConfig(bc, env)
		if err := b.CanExecute(ctx); err != nil {
			clog.Info().Msgf("Skipping because: %v", err)
			continue
		}
		clog.Info().Msg("Executing")
		if err := b.Run(ctx); err != nil {
			return fmt.Errorf("executing backup %q: %w", bc.Name, err)
		}
		state[bc.Name] = env.Clock.Now()
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
