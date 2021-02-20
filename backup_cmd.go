package main

import (
	"fmt"
	"path"

	"github.com/jonboulle/clockwork"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/backup"
	"github.com/mbrt/backsched/internal/config"
	"github.com/mbrt/backsched/internal/exec"
)

const (
	stateFile  = "state.json"
	configFile = "config.jsonnet"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Performs the configured backups",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBackup(); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup() error {
	p := path.Join(cfgDir.Path, configFile)
	cfg, err := config.Parse(p)
	if err != nil {
		return fmt.Errorf("parsing config %q: %w", p, err)
	}
	return backup.Run(ctx, cfg, backup.Env{
		Sio:    stateIO{},
		Clock:  clockwork.NewRealClock(),
		Fs:     afero.NewOsFs(),
		Runner: exec.DefaultRunner{},
	})
}

type stateIO struct{}

func (s stateIO) Save(buf []byte) error {
	if err := cfgDir.WriteFile(stateFile, buf); err != nil {
		statePath := path.Join(cfgDir.Path, stateFile)
		return fmt.Errorf("saving to %q: %v", statePath, err)
	}
	return nil
}

func (s stateIO) Load() ([]byte, error) {
	b, err := cfgDir.ReadFile(stateFile)
	if err != nil {
		statePath := path.Join(cfgDir.Path, stateFile)
		return nil, fmt.Errorf("reading %q: %v", statePath, err)
	}
	return b, nil
}
