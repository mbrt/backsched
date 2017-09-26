package backsched

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNeedBackup(t *testing.T) {
	backup1Time := time.Now().AddDate(0, 0, -10)
	backup2Time := time.Now().AddDate(0, 0, -5)

	state := State{
		"one": backup1Time,
		"two": backup2Time,
	}
	conf := Config{
		Version: "v2",
		Backups: []BackupConf{
			makeTestBackupConf("one", 7),
			makeTestBackupConf("two", 6),
		},
	}
	outdated := ComputeOutdated(conf, state)
	expected := []OutdatedBackup{
		OutdatedBackup{"one", 10, false},
	}
	assert.Equal(t, expected, outdated)
}

func makeTestBackupConf(name string, everyDays int) BackupConf {
	return BackupConf{
		Name:      name,
		EveryDays: everyDays,
		Src:       "",
		SrcDirs:   []string{},
		Rsync:     nil,
		Restic:    nil,
	}
}
