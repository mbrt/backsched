package main

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mbrt/backsched/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runConfig(); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig() error {
	p := path.Join(cfgDir.Path, configFile)
	cfg, err := config.Parse(p)
	if err != nil {
		return fmt.Errorf("parsing config %q: %w", p, err)
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
