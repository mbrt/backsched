package backsched

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	state := State{
		"first": time.Now(),
	}

	// temp file for the serialized state
	file, err := ioutil.TempFile("", "state")
	assert.Nil(t, err)
	fpath := file.Name()
	assert.Nil(t, file.Close())
	defer func() { _ = os.Remove(fpath) }()

	err = state.Save(fpath)
	assert.Nil(t, err)
	s2, err := LoadState(fpath)
	assert.Nil(t, err)

	// purify the timestamps from the crappy monotonic clock reading
	// that would fail the comparison
	for n, t := range state {
		state[n] = t.Round(0)
	}

	assert.Equal(t, state, *s2)
}
