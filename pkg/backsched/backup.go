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
		src, err := ExpandHome(b.Src)
		if err != nil {
			return nil, err
		}
		if b.Rsync != nil {
			dest, err := ExpandHome(b.Rsync.Dest)
			b.Rsync.Dest = dest
			if err != nil {
				return nil, err
			}
			res = append(res, rsyncBackup{src, b.SrcDirs, *b.Rsync})
		}
	}
	return res, nil
}

// BackupResult contains the result of a backup.
type BackupResult struct {
	Name string
	Err  error
}

func (b BackupResult) String() string {
	if b.Err != nil {
		return fmt.Sprintf("%s: %v", b.Name, b.Err)
	}
	return fmt.Sprintf("%s: OK", b.Name)
}

// Backup executes all the backups provided in the config, if possible
func Backup(c Config) error {
	backs, err := MakeBackuppers(c)
	if err != nil {
		return errors.Wrap(err, "error making backuppers from config")
	}
	res := []BackupResult{}
	for i, b := range backs {
		name := c.Backups[i].Name
		if !b.CanBackup() {
			res = append(res, BackupResult{name, errors.New("SKIPPED")})
		} else if err = b.Backup(); err != nil {
			res = append(res, BackupResult{name, errors.Wrap(err, "FAILED")})
		} else {
			res = append(res, BackupResult{name, nil})
		}
	}

	// dump results
	fmt.Printf("\n--- RESULTS ---\n\n")
	for _, r := range res {
		fmt.Printf("  - %v", r)
	}

	return nil
}

type rsyncBackup struct {
	src     string
	srcDirs []string
	rconf   RsyncConf
}

func (r rsyncBackup) Backup() error {
	srcRoot, err := ExpandHome(r.src)
	if err != nil {
		return err
	}
	destRoot, err := ExpandHome(r.rconf.Dest)
	if err != nil {
		return err
	}

	for _, sd := range r.srcDirs {
		spath := EnsureTrailing(path.Join(srcRoot, sd))
		dpath := EnsureTrailing(path.Join(destRoot, sd))
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
