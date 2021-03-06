package main

import (
	"fmt"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/backup"
	"github.com/mbrt/backsched/internal/config"
)

var (
	dryRun     bool
	askSecrets bool
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

	backupCmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "only simulate the backup run.")
	backupCmd.Flags().BoolVarP(&askSecrets, "ask-secrets", "", true, "whether to interactively ask secrets.")
}

func runBackup() error {
	p := path.Join(cfgDir.Path, configFile)
	cfg, err := config.Parse(p)
	if err != nil {
		return fmt.Errorf("parsing config %q: %w", p, err)
	}
	return backup.Run(ctx, cfg, env(), backup.Opts{
		DryRun:     dryRun,
		AskSecrets: askSecrets,
	})
}
