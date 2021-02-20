package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-jsonnet"

	"github.com/mbrt/backsched/internal/errors"
)

// Version is the latest version of the config format.
const Version = "v1alpha1"

// Config is the configuration format.
type Config struct {
	Version string   `json:"version"`
	Backups []Backup `json:"backups"`
}

// Backup is the configuration for a backup.
type Backup struct {
	// Name is the name of the backup. Must be unique.
	Name string `json:"name"`
	// Commands is a list of commands to execute in order.
	Commands []Command `json:"commands"`
	// Requires is an optional list of requirements.
	Requires []Requirement `json:"requires"`
	// Interval is the time interval between backups.
	Interval Duration `json:"interval"`
}

// Command represents a command to run.
type Command struct {
	// Cmd is the full path to the command to run.
	Cmd string `json:"cmd"`
	// Args is the list of arguments to pass
	Args []string `json:"args"`
	// Env is a map of environment variables with their value.
	Env map[string]string `json:"env"`
}

// Requirement is a backup requirement.
type Requirement struct {
	// Path is a path in the filesystem that must be present in order for the
	// backup to proceed.
	Path *string `json:"path"`
}

// Parse takes a file path and returns a parsed config.
func Parse(path string) (Config, error) {
	var cfg Config
	vm := jsonnet.MakeVM()
	js, err := vm.EvaluateFile(path)
	if err != nil {
		return cfg, fmt.Errorf("evaluate jsonnet: %w", err)
	}
	err = jsonUnmarshalStrict([]byte(js), &cfg)
	return cfg, err
}

// Duration is a wrapper arond duration.
//
// See https://github.com/golang/go/issues/10275 and
// https://stackoverflow.com/q/48050945/1667955.
type Duration time.Duration

// MarshalJSON provides custom JSON marshalling for Duration.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON provides custom JSON unmarshalling for Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func jsonUnmarshalStrict(buf []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		// Make the error more informative.
		jctx := contextFromJSONErr(err, buf)
		if jctx == "" {
			return err
		}
		return errors.WithDetails(err,
			fmt.Sprintf("JSON context:\n%s", jctx))
	}
	return nil
}

func contextFromJSONErr(err error, buf []byte) string {
	var (
		jserr  *json.SyntaxError
		juerr  *json.UnmarshalTypeError
		offset int
	)
	switch {
	case errors.As(err, &jserr):
		offset = int(jserr.Offset)
	case errors.As(err, &juerr):
		offset = int(juerr.Offset)
	default:
		return ""
	}

	if offset < 0 || offset >= len(buf) {
		return ""
	}

	// Collect 6 lines of context
	begin, end, count := 0, 0, 0
	for i := offset; i >= 0 && count < 3; i-- {
		if buf[i] == '\n' {
			begin = i + 1
			count++
		}
	}
	for i := offset; i < len(buf) && count < 6; i++ {
		if buf[i] == '\n' {
			end = i
			count++
		}
	}
	return string(buf[begin:end])
}
