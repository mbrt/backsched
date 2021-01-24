package main

import (
	"errors"

	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
)

var (
	cfgDirFlag string
	cfgDirs    configdir.ConfigDir
	cfgDir     *configdir.Config
)

var rootCmd = &cobra.Command{
	Use:   "backsched",
	Short: "Backup scheduler",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgDirFlag, "config", "", "config directory")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfgDirs = configdir.New("mbrt", "backsched")
	cfgDirs.LocalPath = cfgDirFlag
	cfgDir = cfgDirs.QueryFolderContainsFile("config.jsonnet")
	if cfgDir == nil {
		fatal(errors.New(`cannot find config file "config.jsonnet"`))
	}
}
