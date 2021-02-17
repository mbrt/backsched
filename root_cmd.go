package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
)

var (
	// Config.
	cfgDirFlag string
	cfgDirs    configdir.ConfigDir
	cfgDir     *configdir.Config
	cacheDir   *configdir.Config

	// Global context.
	ctx context.Context
)

var rootCmd = &cobra.Command{
	Use:   "backsched",
	Short: "Backup scheduler",
}

func init() {
	cobra.OnInitialize(func() {
		initLogger()
		initCtx()
		initConfig()
	})
	rootCmd.PersistentFlags().StringVar(&cfgDirFlag, "config", "", "config directory")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfgDirs = configdir.New("mbrt", "backsched")
	cfgDirs.LocalPath = cfgDirFlag
	cfgDir = cfgDirs.QueryFolderContainsFile("config.jsonnet")
	if cfgDir == nil {
		log.Fatal().Msgf(`cannot find config file "config.jsonnet"`)
	}
	cacheDir = cfgDirs.QueryCacheFolder()
}

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func initCtx() {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	// Make sure we terminate gracefully on signals by canceling the context.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		s := <-ch
		log.Info().Msgf("Received %v signal: shutting down", s)
		cancel()
	}()
}
