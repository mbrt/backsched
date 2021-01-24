package main

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Performs the configured backups",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBackup(); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup() error {
	p := path.Join(cfgDir.Path, "config.jsonnet")
	cfg, err := Parse(p)
	if err != nil {
		return fmt.Errorf("parsing config %q: %w", p, err)
	}
	fmt.Println(cfg)
	return nil
}
