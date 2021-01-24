package backsched

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testExec struct {
	input     string
	dirExists bool
	execs     []testExecCmd
	mkdirs    []string
}

type testExecCmd struct {
	options ExecOptions
	command []string
}

func (e *testExec) Input(_ string) (string, error) {
	return e.input, nil
}

func (e *testExec) Exec(cmd string, args ...string) error {
	return e.ExecOptions(ExecOptions{"", nil}, cmd, args...)
}

func (e *testExec) ExecOptions(opts ExecOptions, cmd string, args ...string) error {
	cmdArgs := []string{cmd}
	cmdArgs = append(cmdArgs, args...)
	e.execs = append(e.execs, testExecCmd{opts, cmdArgs})
	return nil
}

func (e *testExec) Mkdir(path string, _ os.FileMode) error {
	e.mkdirs = append(e.mkdirs, path)
	return nil
}

func (e testExec) DirExists(path string) bool {
	return e.dirExists
}

func TestRsyncBackup(t *testing.T) {
	conf := Config{
		Version:   "v1",
		StatePath: "",
		Backups: []BackupConf{
			{
				Name:      "foo",
				EveryDays: 3,
				Src:       "/my/src",
				SrcDirs:   []string{"bar/baz", "z"},
				Rsync: &RsyncConf{
					Dest: "/dest",
					Args: []string{"arg1", "arg2"},
				},
			},
		},
	}
	exec := testExec{}
	exec.dirExists = true
	state := State{}
	backs, err := MakeBackuppers(conf)
	assert.Nil(t, err)

	res := backs.Backup(&exec, state)
	expRes := BackupResults{
		BackupResult{
			Name: "foo",
			Err:  nil,
		},
	}
	expCmd := []testExecCmd{
		{ExecOptions{}, []string{"rsync", "arg1", "arg2", "/my/src/bar/baz/", "/dest/bar/baz/"}},
		{ExecOptions{}, []string{"rsync", "arg1", "arg2", "/my/src/z/", "/dest/z/"}},
	}
	expMkdirs := []string{
		"/dest/bar/baz/",
		"/dest/z/",
	}

	assert.Equal(t, expRes, res)
	assert.Equal(t, expCmd, exec.execs)
	assert.Equal(t, expMkdirs, exec.mkdirs)
}

func TestResticBackup(t *testing.T) {
	conf := Config{
		Version:   "v1",
		StatePath: "",
		Backups: []BackupConf{
			{
				Name:      "foo",
				EveryDays: 3,
				Src:       "/my/src",
				SrcDirs:   []string{"bar/baz", "z"},
				Restic: &ResticConf{
					Dest:    ResticDestConf{Dir: "/dest"},
					Check:   false,
					Cleanup: nil,
				},
			},
		},
	}
	exec := testExec{
		input:     "mypass",
		dirExists: true,
		execs:     nil,
		mkdirs:    nil,
	}
	state := State{}
	backs, err := MakeBackuppers(conf)
	assert.Nil(t, err)

	res := backs.Backup(&exec, state)
	expRes := BackupResults{
		BackupResult{
			Name: "foo",
			Err:  nil,
		},
	}
	opts := ExecOptions{
		WorkDir: "/my/src",
		Env:     []string{"RESTIC_PASSWORD=mypass"},
	}
	expCmd := []testExecCmd{
		{opts, []string{"restic", "-r", "/dest", "backup", "bar/baz", "z"}},
	}

	assert.Equal(t, expRes, res)
	assert.Equal(t, expCmd, exec.execs)
	assert.Nil(t, exec.mkdirs)
}

func TestResticBackupCleanupAndCheck(t *testing.T) {
	conf := Config{
		Version:   "v1",
		StatePath: "",
		Backups: []BackupConf{
			{
				Name:      "foo",
				EveryDays: 3,
				Src:       "/my/src",
				SrcDirs:   []string{"bar/baz", "z"},
				Restic: &ResticConf{
					Dest:    ResticDestConf{Dir: "/dest"},
					Check:   true,
					Cleanup: &ResticCleanupConf{4},
				},
			},
		},
	}
	exec := testExec{
		input:     "mypass",
		dirExists: true,
		execs:     nil,
		mkdirs:    nil,
	}
	state := State{}
	backs, err := MakeBackuppers(conf)
	assert.Nil(t, err)

	res := backs.Backup(&exec, state)
	expRes := BackupResults{
		BackupResult{
			Name: "foo",
			Err:  nil,
		},
	}
	opts := ExecOptions{
		WorkDir: "/my/src",
		Env:     []string{"RESTIC_PASSWORD=mypass"},
	}
	expCmd := []testExecCmd{
		{opts, []string{"restic", "-r", "/dest", "backup", "bar/baz", "z"}},
		{opts, []string{"restic", "-r", "/dest", "check"}},
		{opts, []string{"restic", "-r", "/dest", "forget", "--keep-last", "4", "--prune"}},
	}

	assert.Equal(t, expRes, res)
	assert.Equal(t, expCmd, exec.execs)
	assert.Nil(t, exec.mkdirs)
}
