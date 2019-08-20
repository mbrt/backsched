package backsched

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Backup executes all the backups provided in the config, if possible
func Backup(c Config, s State) error {
	exec := osExecutor{}
	backs, err := MakeBackuppers(c)
	if err != nil {
		return errors.Wrap(err, "error making backuppers from config")
	}

	res := backs.Backup(exec, s)
	if err = s.Save(c.StatePath); err != nil {
		fmt.Fprintf(os.Stderr, "error saving state: %v\n", err)
	}

	fmt.Printf("\n--- RESULTS ---\n%v\n", res)

	return err
}

// Backupper is able to perform a certain backup
type Backupper interface {
	// Backup executes the backup
	Backup(ex Executor) error
	// CanBackup returns true if the backup destination is available
	CanBackup(ex Executor) bool
}

// NamedBackupper is a backup executor with a name
type NamedBackupper struct {
	Backupper
	Name string
}

// Backuppers is a list of backup executors
type Backuppers []NamedBackupper

// Backup execute the backup for all possible backuppers
func (bs Backuppers) Backup(exec Executor, state State) BackupResults {
	res := []BackupResult{}
	for _, b := range bs {
		if !b.CanBackup(exec) {
			res = append(res, BackupResult{b.Name, errors.New("SKIPPED")})
			continue
		}
		if err := b.Backup(exec); err != nil {
			res = append(res, BackupResult{b.Name, errors.Wrap(err, "FAILED")})
			continue
		}
		res = append(res, BackupResult{b.Name, nil})
		state[b.Name] = time.Now()
	}
	return res
}

// MakeBackuppers creates a list of backuppers from a config.
func MakeBackuppers(c Config) (Backuppers, error) {
	var res Backuppers
	for _, b := range c.Backups {
		if b.Rsync != nil {
			rb := rsyncBackup{b.Src, b.SrcDirs, *b.Rsync}
			res = append(res, NamedBackupper{rb, b.Name})
		}
		if b.Restic != nil {
			rb := resticBackup{b.Src, b.SrcDirs, *b.Restic}
			res = append(res, NamedBackupper{rb, b.Name})
		}
	}
	return res, nil
}

// BackupResults is a list of backup results.
type BackupResults []BackupResult

func (bs BackupResults) String() string {
	res := make([]string, len(bs))
	for i, r := range bs {
		res[i] = fmt.Sprintf("  - %v", r)
	}
	return strings.Join(res, "\n")
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
	for _, sd := range r.srcDirs {
		spath := EnsureTrailing(path.Join(r.src, sd))
		dpath := EnsureTrailing(path.Join(r.rconf.Dest, sd))
		if err := ex.Mkdir(dpath, 0755); err != nil {
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
	return ex.DirExists(r.src) && ex.DirExists(r.rconf.Dest)
}

type resticBackup struct {
	src     string
	srcDirs []string
	rconf   ResticConf
}

func (r resticBackup) Backup(ex Executor) error {
	pwd, err := ex.Input(fmt.Sprintf("[%s] backup password", r.dest()))
	if err != nil {
		return errors.Wrap(err, "error in password input")
	}

	opts := ExecOptions{
		WorkDir: r.src,
		Env:     r.env(pwd),
	}

	args := []string{"-r", r.dest(), "backup"}
	args = append(args, r.srcDirs...)
	if err := ex.ExecOptions(opts, "restic", args...); err != nil {
		return errors.Wrap(err, "restic backup failed")
	}

	if r.rconf.Check {
		args = []string{"-r", r.dest(), "check"}
		if err := ex.ExecOptions(opts, "restic", args...); err != nil {
			return errors.Wrap(err, "restic check failed")
		}
	}

	if r.rconf.Cleanup != nil {
		keepLast := r.rconf.Cleanup.KeepLast
		if keepLast <= 0 {
			return errors.New("restic.cleanup.keepLast > 0 is required")
		}
		args = []string{"-r", r.dest(), "forget", "--keep-last", strconv.Itoa(keepLast), "--prune"}
		if err := ex.ExecOptions(opts, "restic", args...); err != nil {
			return errors.Wrap(err, "restic cleanup failed")
		}
	}

	return nil
}

func (r resticBackup) CanBackup(ex Executor) bool {
	if !ex.DirExists(r.src) {
		return false
	}
	if r.rconf.Dest.GCloud != nil {
		// No way to know whether we can backup remotely. Assuming we can.
		return true
	}
	return ex.DirExists(r.rconf.Dest.Dir)
}

func (r resticBackup) dest() string {
	if r.rconf.Dest.GCloud != nil {
		return r.rconf.Dest.GCloud.Bucket
	}
	return r.rconf.Dest.Dir
}

func (r resticBackup) env(pwd string) []string {
	res := []string{fmt.Sprintf("RESTIC_PASSWORD=%s", pwd)}
	if gcloud := r.rconf.Dest.GCloud; gcloud != nil {
		res = append(res, fmt.Sprintf("GOOGLE_PROJECT_ID=%s", gcloud.ProjectID))
		res = append(res, fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", gcloud.CredPath))
	}
	return res
}
