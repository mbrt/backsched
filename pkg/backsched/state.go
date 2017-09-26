package backsched

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// LoadState loads the backup state from a file
func LoadState(path string) (*State, error) {
	s, err := loadState(path)
	if err != nil {
		return nil, err
	}
	res := fromMarshalState(*s)
	return &res, nil
}

// State contains the state of all the backups
type State map[string]time.Time

// LastBackupOf returns the time of the last given backup name.
// Returns false if the backup has been never done
func (s State) LastBackupOf(name string) (time.Time, bool) {
	if t, ok := s[name]; ok {
		return t, true
	}
	return time.Time{}, false
}

// Save saves the state to the given file
func (s State) Save(path string) error {
	return s.toMarshal().Save(path)
}

func (s State) toMarshal() marshalState {
	backups := make([]marshalBackupState, len(s))
	i := 0
	for k, v := range s {
		backups[i] = marshalBackupState{k, v}
		i++
	}
	return marshalState{"v1", backups}
}

type marshalState struct {
	Version string
	Backups []marshalBackupState
}

// BackupState contains the state of one backup
type marshalBackupState struct {
	Name       string
	LastBackup time.Time `yaml:"lastBackup"`
}

func fromMarshalState(s marshalState) State {
	res := State{}
	for _, b := range s.Backups {
		res[b.Name] = b.LastBackup
	}
	return res
}

// LoadState loads the backup state from a file
func loadState(path string) (*marshalState, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open state file")
	}
	defer func() { _ = file.Close() }()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "corrupted state file")
	}
	var res marshalState
	err = yaml.Unmarshal(contents, &res)
	if err != nil {
		return nil, errors.Wrap(err, "invalid state file")
	}
	return &res, nil
}

// Save saves the state to the given file
func (s marshalState) Save(path string) (err error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "cannot open state file for write")
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	bytes, err := yaml.Marshal(&s)
	if err != nil {
		return errors.Wrap(err, "cannot marshal the state")
	}
	_, err = file.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "cannot write to the state file")
	}
	return err
}
