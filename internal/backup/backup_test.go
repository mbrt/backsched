package backup_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbrt/backsched/internal/backup"
	"github.com/mbrt/backsched/internal/config"
	"github.com/mbrt/backsched/internal/exec"
)

// testRunner is a fake exec.Runner.
//
// This struct keeps track of what is ran by creating json files by using the
// given Fs, under /run/<num>. The files are numbered incrementally and
// contain a JSON representation of the Run arguments.
type testRunner struct {
	fs    afero.Fs
	count int
}

func (t *testRunner) Run(ctx context.Context, cmd exec.Cmd) error {
	t.count++
	if err := t.fs.MkdirAll("/run", 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cmd, "", "    ")
	if err != nil {
		return err
	}
	return afero.WriteFile(t.fs, fmt.Sprintf("/run/%d", t.count), data, 0o600)
}

type testSio struct {
	fs afero.Fs
}

func (t testSio) Save(buf []byte) error {
	return afero.WriteFile(t.fs, "/state", buf, 0o600)
}

func (t testSio) Load() ([]byte, error) {
	return afero.ReadFile(t.fs, "/state")
}

type testSecrets struct{}

func (testSecrets) Secret(backup string, s config.Secret) string {
	return fmt.Sprintf("%s-%s-val", backup, s.ID)
}

func loadTestFile(t *testing.T, p string) []byte {
	t.Helper()
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func checkFile(t *testing.T, fs afero.Fs, path string, expected []byte) {
	t.Helper()
	got, err := afero.ReadFile(fs, path)
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestComplete(t *testing.T) {
	// Load test files.
	weekly := [][]byte{
		loadTestFile(t, "testfiles/run_weekly_0.json"),
		loadTestFile(t, "testfiles/run_weekly_1.json"),
	}
	hourly := [][]byte{
		loadTestFile(t, "testfiles/run_hourly_0.json"),
		loadTestFile(t, "testfiles/run_hourly_1.json"),
	}
	cfg, err := config.Parse("testfiles/complete.jsonnet")
	require.Nil(t, err)

	// Set up fake environment.
	ctx := context.Background()
	clock := clockwork.NewFakeClock()
	fs := afero.NewMemMapFs()
	runner := testRunner{fs, 0}
	sio := testSio{fs}
	env := backup.Env{
		Clock:   clock,
		Fs:      fs,
		Runner:  &runner,
		Sio:     sio,
		Secrets: testSecrets{},
	}
	opts := backup.Opts{
		DryRun:     false,
		AskSecrets: true,
	}

	// Run without requirements. Nothing should execute.
	err = backup.Run(ctx, cfg, env, opts)
	require.Nil(t, err)
	ok, _ := afero.Exists(fs, "/run/0")
	assert.Equal(t, false, ok)
	// Everything is outdated.
	od, err := backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Equal(t, []backup.Info{
		{Since: 0, Backup: cfg.Backups[0]},
		{Since: 0, Backup: cfg.Backups[1]},
	}, od)

	// Add the required paths for hourly only.
	err = fs.MkdirAll("/mnt/backup/dir2", 0x700)
	require.Nil(t, err)

	// Run the backup: only hourly should be executed.
	err = backup.Run(ctx, cfg, env, opts)
	require.Nil(t, err)
	// Check the commands that ran.
	checkFile(t, fs, "/run/1", hourly[0])
	checkFile(t, fs, "/run/2", hourly[1])
	// Weekly shouldn't have ran.
	ok, _ = afero.Exists(fs, "/run/3")
	assert.Equal(t, false, ok)
	// Only weekly is now outdated.
	od, err = backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Equal(t, []backup.Info{
		{Since: 0, Backup: cfg.Backups[0]},
	}, od)

	// Create all required paths, but run too early for hourly.
	clock.Advance(30 * time.Minute)
	err = fs.MkdirAll("/mnt/backup/dir1", 0x700)
	require.Nil(t, err)
	// Run the backup.
	err = backup.Run(ctx, cfg, env, opts)
	require.Nil(t, err)
	// Check the commands that ran.
	checkFile(t, fs, "/run/3", weekly[0])
	checkFile(t, fs, "/run/4", weekly[1])
	// Hourly shouldn't have ran.
	ok, _ = afero.Exists(fs, "/run/5")
	assert.Equal(t, false, ok)
	// Nothing is outdated.
	od, err = backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Equal(t, []backup.Info(nil), od)

	// Wait more, now hourly can run but weekly can't.
	clock.Advance(45 * time.Minute)
	// Hourly is outdated.
	od, err = backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Equal(t, []backup.Info{
		{Since: 75 * time.Minute, Backup: cfg.Backups[1]},
	}, od)
	// Run the backup.
	err = backup.Run(ctx, cfg, env, opts)
	require.Nil(t, err)
	// Check the commands that ran.
	checkFile(t, fs, "/run/5", hourly[0])
	checkFile(t, fs, "/run/6", hourly[1])
	// Weekly shouldn't have ran.
	ok, _ = afero.Exists(fs, "/run/7")
	assert.Equal(t, false, ok)

	// Now wait a week. Everything should be ready to run.
	clock.Advance(8 * 24 * time.Hour)
	// Everything is outdated.
	od, err = backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Equal(t, []backup.Info{
		{
			Since:  8*24*time.Hour + 45*time.Minute,
			Backup: cfg.Backups[0],
		},
		{
			Since:  8 * 24 * time.Hour,
			Backup: cfg.Backups[1],
		},
	}, od)
	// Run the backup.
	err = backup.Run(ctx, cfg, env, opts)
	require.Nil(t, err)
	// Check the commands that ran.
	checkFile(t, fs, "/run/7", weekly[0])
	checkFile(t, fs, "/run/8", weekly[1])
	checkFile(t, fs, "/run/9", hourly[0])
	checkFile(t, fs, "/run/10", hourly[1])
}

type faultySecrets struct {
	t *testing.T
}

func (f faultySecrets) Secret(backup string, s config.Secret) string {
	f.t.Fatalf("Backup %s asked for secret %s, expected no call", backup, s.ID)
	return ""
}

func TestSecrets(t *testing.T) {
	cfg, err := config.Parse("testfiles/complete.jsonnet")
	require.Nil(t, err)

	// Set up fake environment.
	ctx := context.Background()
	clock := clockwork.NewFakeClock()
	fs := afero.NewMemMapFs()
	runner := testRunner{fs, 0}
	sio := testSio{fs}
	env := backup.Env{
		Clock:   clock,
		Fs:      fs,
		Runner:  &runner,
		Sio:     sio,
		Secrets: faultySecrets{t},
	}

	// Add required paths.
	err = fs.MkdirAll("/mnt/backup/dir1", 0x700)
	assert.Nil(t, err)
	err = fs.MkdirAll("/mnt/backup/dir2", 0x700)
	assert.Nil(t, err)

	// Both backups should be outdated.
	od, err := backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Len(t, od, 2)

	// Backup should run fine, but no secrets should be asked.
	err = backup.Run(ctx, cfg, env, backup.Opts{
		DryRun:     false,
		AskSecrets: false,
	})
	assert.Nil(t, err)

	// Backups are now fine.
	od, err = backup.ComputeOutdated(ctx, cfg, env)
	assert.Nil(t, err)
	assert.Empty(t, od)

	// Let enough time pass for the backups to be outdated again.
	clock.Advance(10 * 24 * time.Hour)

	// Dry run doesn't ask for secrets either.
	err = backup.Run(ctx, cfg, env, backup.Opts{
		DryRun:     true,
		AskSecrets: true,
	})
	assert.Nil(t, err)
}
