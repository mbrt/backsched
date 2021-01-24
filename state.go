package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// LoadState loads the backup state from a file
func LoadState(path string) (State, error) {
	f, err := os.Open(path)
	if err != nil {
		return State{}, err
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return State{}, fmt.Errorf("reading state %q: %v", path, err)
	}
	var res State
	err = jsonUnmarshalStrict(buf, &res)
	return res, nil
}

// State contains the state of all the backups
type State map[string]time.Time

// LastBackupOf returns the time of the last given backup name.
// Returns false if the backup has been never done
func (s State) LastBackupOf(name string) (time.Time, bool) {
	if t, ok := s[name]; ok {
		return t, true
	}
	return time.Time{}, false
}

// Save saves the state to the given file
func (s State) Save(path string) error {
	buf, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", path, err)
	}
	defer f.Close()
	_, err = f.Write(buf)
	return err
}
