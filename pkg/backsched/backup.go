package backsched

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
)

// Backup executes all the backups provided in the config, if possible
func Backup(c Config, s State) error {
	backs, err := MakeBackuppers(c)
	if err != nil {
		return errors.Wrap(err, "error making backuppers from config")
	}
	exec := cmdExecutor{}
	res := []BackupResult{}
	for i, b := range backs {
		name := c.Backups[i].Name
		if !b.CanBackup(exec) {
			res = append(res, BackupResult{name, errors.New("SKIPPED")})
			continue
		}
		if err = b.Backup(exec); err != nil {
			res = append(res, BackupResult{name, errors.Wrap(err, "FAILED")})
			continue
		}
		res = append(res, BackupResult{name, nil})
		s[name] = time.Now()
	}

	if err = s.Save(c.StatePath); err != nil {
		fmt.Fprintf(os.Stderr, "error saving state: %v\n", err)
	}

	// dump results
	fmt.Printf("\n--- RESULTS ---\n")
	for _, r := range res {
		fmt.Printf("  - %v\n", r)
	}

	return err
}

// Backupper is able to perform a certain backup
type Backupper interface {
	// Backup executes the backup
	Backup(ex Executor) error
	// CanBackup returns true if the backup destination is available
	CanBackup(ex Executor) bool
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

type rsyncBackup struct {
	src     string
	srcDirs []string
	rconf   RsyncConf
}

func (r rsyncBackup) Backup(ex Executor) error {
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
		if err := ex.Exec("rsync", args...); err != nil {
			return errors.Wrap(err, "rsync failed")
		}
	}
	return nil
}

func (r rsyncBackup) CanBackup(ex Executor) bool {
	return dirExists(r.src) && dirExists(r.rconf.Dest)
}

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
