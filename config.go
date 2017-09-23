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
	Version string       `yaml:"version"`
	Backups []BackupConf `yaml:"backups"`
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
	Name      string      `yaml:"name"`
	EveryDays int         `yaml:"everyDays"`
	Src       string      `yaml:"src"`
	SrcDirs   []string    `yaml:"srcDirs"`
	Rsync     *RsyncConf  `yaml:"rsync"`
	Restic    *ResticConf `yamlm:"restic"`
}

// RsyncConf contains the rsync configuration for one backup
type RsyncConf struct {
	Dest string   `yaml:"dest"`
	Args []string `yaml:"args"`
}

// ResticConf is the restic configuration for one backup
type ResticConf struct {
	Dest    string            `yaml:"dest"`
	Check   bool              `yaml:"check"`
	Cleanup ResticCleanupConf `yaml:"cleanup"`
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
