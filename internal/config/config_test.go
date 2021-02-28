package config_test

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mbrt/backsched/internal/config"
)

var update = flag.Bool("update", false, "Update golden files.")

func paths(t *testing.T, pattern string) []string {
	t.Helper()
	res, err := filepath.Glob(pattern)
	require.Nil(t, err)
	return res
}

func dumpJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	require.Nil(t, err)
	f, err := os.Create(path)
	require.Nil(t, err)
	defer f.Close()
	_, err = f.Write(b)
	require.Nil(t, err)
}

func parseJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	f, err := os.Open(path)
	require.Nil(t, err)
	b, err := io.ReadAll(f)
	require.Nil(t, err)
	err = json.Unmarshal(b, v)
	require.Nil(t, err)
}

func TestParse(t *testing.T) {
	tpaths := paths(t, "testfiles/ok-*.jsonnet")
	for _, tp := range tpaths {
		t.Run(tp, func(t *testing.T) {
			cfg, err := config.Parse(tp)
			require.Nil(t, err)

			goldenp := strings.Replace(tp, ".jsonnet", ".json", 1)
			if *update {
				// Update golden files only.
				dumpJSON(t, goldenp, cfg)
				return
			}

			var golden config.Config
			parseJSON(t, goldenp, &golden)
			assert.Equal(t, golden, cfg)
		})
	}
}
