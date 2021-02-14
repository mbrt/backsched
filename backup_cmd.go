package main

import (
	"context"
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/exec"
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
	ctx := context.Background()

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
	}

	return nil
}

func executorFromCfg(bc BackupCfg) exec.Executor {
	var reqs []exec.Req
	for _, r := range bc.Requires {
		if r.Path != nil {
			reqs = append(reqs, exec.DirExists{Path: *r.Path})
		}
	}

	// TODO: Handle bc.SecretEnv.
	return exec.Executor{
		Cfg: exec.Config{
			Reqs: reqs,
			Cmds: []exec.Cmd{
				{
					Exe:  bc.Cmd,
					Env:  bc.Env,
					Args: bc.Args,
				},
			},
		},
		Os: exec.LocalOs{},
	}
}
