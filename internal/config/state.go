package config

import (
	"encoding/json"
	"time"
)

// LoadState loads the backup state from a buffer
func LoadState(b []byte) (State, error) {
	var res State
	err := jsonUnmarshalStrict(b, &res)
	return res, err
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
func (s State) Save() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
