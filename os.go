package main

import (
	"fmt"
	"path"

	"github.com/jonboulle/clockwork"
	"github.com/spf13/afero"

	"github.com/mbrt/backsched/internal/backup"
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

func env() backup.Env {
	return backup.Env{
		Sio:    stateIO{},
		Clock:  clockwork.NewRealClock(),
		Fs:     afero.NewOsFs(),
		Runner: exec.DefaultRunner{},
	}
}
