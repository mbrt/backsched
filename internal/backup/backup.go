package backup

import (
	"context"
	"fmt"

	"github.com/jonboulle/clockwork"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"

	"github.com/mbrt/backsched/internal/config"
	"github.com/mbrt/backsched/internal/exec"
)

// Run runs the given backup configuration.
func Run(ctx context.Context, cfg config.Config, env Env, opts Opts) error {
	backups, err := ComputeOutdated(ctx, cfg, env)
	if err != nil {
		return err
	}

	state := config.State{}
	if opts.DryRun {
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

		b := newExecutorFromConfig(bc.Backup, env, opts)
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

// Opts groups contains backup options.
type Opts struct {
	DryRun     bool
	AskSecrets bool
}

// Env groups together the environment a backup is ran against.
type Env struct {
	Sio     StateIOer
	Clock   clockwork.Clock
	Fs      afero.Fs
	Runner  exec.Runner
	Secrets SecretGetter
}

// StateIOer abstracts away lower level save and load functionality for the
// backup state.
type StateIOer interface {
	Save(buf []byte) error
	Load() ([]byte, error)
}

// SecretGetter returns the value of a secret.
type SecretGetter interface {
	Secret(backup string, s config.Secret) string
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

func newExecutorFromConfig(bc config.Backup, env Env, opts Opts) exec.Executor {
	var reqs []exec.Requirement
	for _, r := range bc.Requires {
		if r.Path != nil {
			reqs = append(reqs, exec.DirExists{Path: *r.Path})
		}
	}

	secrets := map[string]string{}
	if opts.AskSecrets && !opts.DryRun {
		secrets = collectSecretVals(bc, env.Secrets)
	}

	var cmds []exec.Cmd
	for _, c := range bc.Commands {
		// Map the secret environment to the already collected secret values.
		// The match is done by ID.
		secEnv := map[string]string{}
		for env, sec := range c.SecretEnv {
			// It's possible the secret is not present, if we didn't ask them
			// in the first place.
			if val, ok := secrets[sec.ID]; ok {
				secEnv[env] = val
			}
		}
		cmds = append(cmds, exec.Cmd{
			Cmd:       c.Cmd,
			Args:      c.Args,
			Env:       c.Env,
			Workdir:   c.Workdir,
			SecretEnv: secEnv,
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

func collectSecretVals(bc config.Backup, sg SecretGetter) map[string]string {
	res := map[string]string{}
	for _, cmd := range bc.Commands {
		for _, s := range cmd.SecretEnv {
			if _, ok := res[s.ID]; ok {
				// We already know about this secret.
				continue
			}
			res[s.ID] = sg.Secret(bc.Name, s)
		}
	}
	return res
}

type dryRunner struct{}

// Run runs a command as a subprocess.
func (dryRunner) Run(ctx context.Context, cmd exec.Cmd) error {
	log.Info().
		Str("cmd", cmd.Cmd).
		Str("args", fmt.Sprintf("%v", cmd.Args)).
		Str("env", fmt.Sprintf("%v", cmd.Env)).
		Str("secrets", fmt.Sprintf("%v", keys(cmd.SecretEnv))).
		Msg("Would have run the command")
	return nil
}

func keys(m map[string]string) []string {
	var res []string
	for k := range m {
		res = append(res, k)
	}
	return res
}
