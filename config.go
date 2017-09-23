package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config contains the parsed config file
type Config struct {
	Version string
	Backups []BackupConf
}

func (c Config) String() string {
	r, err := yaml.Marshal(&c)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(r)
}

// BackupConf contains the configuration of one backup
type BackupConf struct {
	Name      string
	EveryDays int
	Src       string
	SrcDirs   []string
	Rsync     *RsyncConf
	Restic    *ResticConf
}

// RsyncConf contains the rsync configuration for one backup
type RsyncConf struct {
	Dest string
	Args []string
}

// ResticConf is the restic configuration for one backup
type ResticConf struct {
	Dest    string
	Check   bool
	Cleanup ResticCleanupConf
}

// ResticCleanupConf is the restic cleanup configuration for one backup.
type ResticCleanupConf struct {
	KeepLast int `yaml:"keepLast"`
}

// ParseConfig parses the given config file path.
func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config file")
	}
	defer func() { _ = file.Close() }()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "corrupted config file")
	}

	var res Config
	err = yaml.Unmarshal(contents, &res)
	if err != nil {
		return nil, errors.Wrap(err, "invalid config file")
	}
	return &res, nil
}
