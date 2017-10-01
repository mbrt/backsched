package backsched

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testExec struct {
	dirExists bool
	execs     [][]string
	mkdirs    []string
}

func (e *testExec) Exec(cmd string, args ...string) error {
	cmdArgs := []string{cmd}
	cmdArgs = append(cmdArgs, args...)
	e.execs = append(e.execs, cmdArgs)
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
			BackupConf{
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
	expCmd := [][]string{
		[]string{"rsync", "arg1", "arg2", "/my/src/bar/baz/", "/dest/bar/baz/"},
		[]string{"rsync", "arg1", "arg2", "/my/src/z/", "/dest/z/"},
	}
	expMkdirs := []string{
		"/dest/bar/baz/",
		"/dest/z/",
	}

	assert.Equal(t, expRes, res)
	assert.Equal(t, expCmd, exec.execs)
	assert.Equal(t, expMkdirs, exec.mkdirs)
}
