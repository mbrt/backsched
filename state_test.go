package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	state := State{
		Version: "v1",
		Backups: []BackupState{
			BackupState{Name: "first", LastBackup: time.Now()},
		},
	}

	// temp file for the serialized state
	file, err := ioutil.TempFile("", "state")
	assert.Nil(t, err)
	fpath := file.Name()
	assert.Nil(t, file.Close())
	defer func() { _ = os.Remove(fpath) }()

	err = state.Save(fpath)
	assert.Nil(t, err)
	s2, err := LoadState(fpath)
	assert.Nil(t, err)

	// purify the timestamps from the crappy monotonic clock reading
	// that would fail the comparison
	for i, b := range state.Backups {
		t := b.LastBackup.Round(0)
		state.Backups[i].LastBackup = t
	}

	assert.Equal(t, state, *s2)
}
