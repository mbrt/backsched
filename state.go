package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// State contains the state of all the backups
type State struct {
	Version string
	Backups []BackupState
}

// BackupState contains the state of one backup
type BackupState struct {
	Name       string
	LastBackup time.Time
}

// LoadState loads the backup state from a file
func LoadState(path string) (*State, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open state file")
	}
	defer func() { _ = file.Close() }()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "corrupted state file")
	}
	var res State
	err = yaml.Unmarshal(contents, &res)
	if err != nil {
		return nil, errors.Wrap(err, "invalid state file")
	}
	return &res, nil
}

// Save saves the state to the given file
func (s State) Save(path string) (err error) {
	file, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrap(err, "cannot open state file for write")
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	bytes, err := yaml.Marshal(&s)
	if err != nil {
		return errors.Wrap(err, "cannot marshal the state")
	}
	_, err = file.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "cannot write to the state file")
	}
	return err
}
