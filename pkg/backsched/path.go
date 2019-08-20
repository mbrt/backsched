package backsched

import (
	"os/user"
	"path"

	"github.com/pkg/errors"
)

// ExpandHome expands a path to the current home, if it starts with '~/'
func ExpandHome(p string) (string, error) {
	if len(p) == 0 || p[0] != '~' {
		return p, nil
	}
	if len(p) > 1 && p[:2] != "~/" {
		return p, nil
	}

	var rest string
	if len(p) == 1 {
		rest = ""
	} else {
		rest = p[2:]
	}

	home, err := Home()
	if err != nil {
		return "", errors.Wrap(err, "cannot expand home dir in path")
	}
	p = path.Join(home, rest)
	return p, nil
}

// Home returns the current user home directory.
func Home() (string, error) {
	usr, err := user.Current()
	return usr.HomeDir, err
}

// EnsureTrailing ensures that the given path ends with a trailing slash.
func EnsureTrailing(p string) string {
	if p[len(p)-1] == '/' {
		return p
	}
	return p + "/"
}
