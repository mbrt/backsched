package main

import (
	"fmt"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/backup"
	"github.com/mbrt/backsched/internal/config"
)

const stateFile = "state.json"

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Performs the configured backups",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBackup(); err != nil {
			log.Fatal().Err(err).Msg("Run backup command")
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
	return backup.Run(ctx, cfg, stateIO{})
}

type stateIO struct{}

func (s stateIO) Save(buf []byte) error {
	if err := cacheDir.WriteFile(stateFile, buf); err != nil {
		statePath := path.Join(cacheDir.Path, stateFile)
		return fmt.Errorf("saving to %q: %v", statePath, err)
	}
	return nil
}

func (s stateIO) Load() ([]byte, error) {
	b, err := cacheDir.ReadFile(stateFile)
	if err != nil {
		statePath := path.Join(cacheDir.Path, stateFile)
		return nil, fmt.Errorf("reading %q: %v", statePath, err)
	}
	return b, nil
}
