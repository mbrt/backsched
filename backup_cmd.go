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

	for _, bc := range cfg.Backups {
		b := NewBackup(bc, OSExecutor{}, nil)
		if err := b.CanExecute(); err != nil {
			fmt.Printf("Skipping backup %q because: %v", b.cfg.Name, err)
			continue
		}
		fmt.Printf("Executing backup %q\n", bc.Name)
		if err := b.Run(); err != nil {
			return fmt.Errorf("backup %q failed: %w", bc.Name, err)
		}
	}

	return nil
}
