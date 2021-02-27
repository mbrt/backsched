package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/backup"
	"github.com/mbrt/backsched/internal/config"
)

var notify bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the state of the configured backups",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCheck(); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().BoolVarP(&notify, "notify", "", false, "whether to send desktop notifications.")
}

func runCheck() error {
	p := path.Join(cfgDir.Path, configFile)
	cfg, err := config.Parse(p)
	if err != nil {
		return fmt.Errorf("parsing config %q: %w", p, err)
	}
	infos, err := backup.ComputeOutdated(ctx, cfg, env())
	if err != nil {
		return fmt.Errorf("in compute outdated: %v", err)
	}

	for _, b := range infos {
		log.Info().Str("backup", b.Backup.Name).Msgf("Needs backup: %v", b)
	}
	if notify && len(infos) > 0 {
		return report(infos)
	}

	return nil
}

func report(infos []backup.Info) error {
	summary := "The following backups are outdated:"
	var msg []string
	for _, info := range infos {
		msg = append(msg, fmt.Sprintf("  - %s\n", info))
	}
	return beeep.Notify(summary, strings.Join(msg, ""), "")
}
