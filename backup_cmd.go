package main

import (
	"fmt"
	"path"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/config"
	"github.com/mbrt/backsched/internal/exec"
)

const stateFile = "state.json"

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Performs the configured backups",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBackup(); err != nil {
			log.Fatal().Err(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup() error {
	p := path.Join(cfgDir.Path, "config.jsonnet")
	cfg, err := config.Parse(p)
	if err != nil {
		return fmt.Errorf("parsing config %q: %w", p, err)
	}

	state := loadState()
	// Make sure we save the state at the end.
	defer saveState(state)

	for _, bc := range cfg.Backups {
		b := executorFromCfg(bc)
		if err := b.CanExecute(ctx); err != nil {
			fmt.Printf("Skipping backup %q because: %v", bc.Name, err)
			continue
		}
		fmt.Printf("Executing backup %q\n", bc.Name)
		if err := b.Run(ctx); err != nil {
			return fmt.Errorf("backup %q failed: %w", bc.Name, err)
		}
		state[bc.Name] = time.Now()
	}

	return nil
}

func loadState() config.State {
	statePath := path.Join(cacheDir.Path, stateFile)
	b, err := cacheDir.ReadFile(stateFile)
	if err != nil {
		log.Warn().Msgf("State file %q not found", statePath)
		return config.State{}
	}
	state, err := config.LoadState(b)
	if err != nil {
		log.Error().Err(err).Msgf("Parsing state file %q", statePath)
		return config.State{}
	}
	return state
}

func saveState(s config.State) {
	statePath := path.Join(cacheDir.Path, stateFile)
	buf, err := s.Save()
	if err != nil {
		log.Error().Err(err).Msgf("Serializing state file")
	}
	if err := cacheDir.WriteFile(stateFile, buf); err != nil {
		log.Error().Err(err).Msgf("Writing state file %q", statePath)
	}
}

func executorFromCfg(bc config.Backup) exec.Executor {
	var reqs []exec.Req
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
		Os: exec.LocalOs{},
	}
}
