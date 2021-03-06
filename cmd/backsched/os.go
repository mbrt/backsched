package main

import (
	"fmt"
	"os"
	"path"

	"github.com/jonboulle/clockwork"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"golang.org/x/term"

	"github.com/mbrt/backsched/internal/backup"
	"github.com/mbrt/backsched/internal/config"
	"github.com/mbrt/backsched/internal/exec"
)

const (
	stateFile  = "state.json"
	configFile = "config.jsonnet"
)

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

type secrets struct{}

func (secrets) Secret(backup string, s config.Secret) string {
	fmt.Printf("[backup %q %s]: ", backup, s.ID)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal().
			Err(err).
			Str("backup", backup).
			Msgf("Reading secret for ID=%s", s.ID)
	}
	return string(b)
}

func env() backup.Env {
	return backup.Env{
		Sio:     stateIO{},
		Clock:   clockwork.NewRealClock(),
		Fs:      afero.NewOsFs(),
		Runner:  exec.DefaultRunner{},
		Secrets: secrets{},
	}
}
