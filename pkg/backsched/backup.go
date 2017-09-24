package backsched

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
)

// Backupper is able to perform a certain backup
type Backupper interface {
	// Backup executes the backup
	Backup() error
	// CanBackup returns true if the backup destination is available
	CanBackup() bool
}

// MakeBackuppers creates a list of backuppers from a config.
func MakeBackuppers(c Config) ([]Backupper, error) {
	var res []Backupper
	for _, b := range c.Backups {
		if b.Rsync != nil {
			res = append(res, rsyncBackup{b.Src, b.SrcDirs, *b.Rsync})
		}
	}
	return res, nil
}

type rsyncBackup struct {
	src     string
	srcDirs []string
	rconf   RsyncConf
}

func (r rsyncBackup) Backup() error {
	for _, sd := range r.srcDirs {
		spath := path.Join(r.src, sd)
		dpath := path.Join(r.rconf.Dest, sd)
		if err := os.MkdirAll(dpath, 0755); err != nil {
			return errors.Wrap(err, fmt.Sprintf("cannot create dest %v", dpath))
		}
		args := []string{}
		args = append(args, r.rconf.Args...)
		args = append(args, spath, dpath)
		sp := exec.Command("rsync", args...)
		if err := sp.Run(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("errors executing backup %v", dpath))
		}
	}
	return nil
}

func (r rsyncBackup) CanBackup() bool {
	return dirExists(r.src) && dirExists(r.rconf.Dest)
}

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err != nil && stat.IsDir()
}
