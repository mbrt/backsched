package backsched

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config contains the parsed config file
type Config struct {
	Version   string
	StatePath string `yaml:"statePath"`
	Backups   []BackupConf
}

func (c Config) String() string {
	r, err := yaml.Marshal(&c)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(r)
}

func (c *Config) expandPaths() {
	c.StatePath, _ = ExpandHome(c.StatePath)
	for i := range c.Backups {
		b := &c.Backups[i]
		b.Src, _ = ExpandHome(b.Src)
		// no need to expand the source dirs
		if b.Rsync != nil {
			b.Rsync.Dest, _ = ExpandHome(b.Rsync.Dest)
		}
		if b.Restic != nil {
			b.Restic.Dest, _ = ExpandHome(b.Restic.Dest)
		}
	}
}

// BackupConf contains the configuration of one backup
type BackupConf struct {
	Name      string
	EveryDays int `yaml:"everyDays"`
	Src       string
	SrcDirs   []string    `yaml:"srcDirs"`
	Rsync     *RsyncConf  `yaml:",omitempty"`
	Restic    *ResticConf `yaml:",omitempty"`
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
	KeepLast int
}

// ParseConfig parses the given config file path.
func ParseConfig(path string) (*Config, error) {
	path, err := ExpandHome(path)
	if err != nil {
		return nil, err
	}

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
	res.expandPaths()
	return &res, nil
}
