package main

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/mbrt/backsched/pkg/backsched"
)

type commandType int
type logLevel int

const (
	cmdNeedBackup = iota
	cmdDoBackup

	logVerbose = iota
	logNormal

	defaultConfFile  = "~/.backsched/config.yaml"
	defaultStateFile = "~/.backsched/state.yaml"
)

type cmdlineOpts struct {
	command   commandType
	logs      logLevel
	confPath  string
	statePath string
}

func parseArgs() cmdlineOpts {
	usage := `backsched.

Usage:
  backsched need-backup [options]
  backsched backup [options]

Options:
  -h --help               Show this screen.
  --verbose               Verbose output.
  --config=<conf-file>    Use the specified config file.
  --state=<state-file>    Use the spefied state file.`

	args, _ := docopt.Parse(usage, nil, true, "Backup Scheduler 0.1", false)

	result := cmdlineOpts{
		cmdDoBackup,
		logNormal,
		defaultConfFile,
		defaultStateFile,
	}

	if args["need-backup"] != nil {
		result.command = cmdNeedBackup
	}
	if args["--verbose"].(bool) {
		result.logs = logVerbose
	}
	if conf, ok := args["--config"].(string); ok {
		result.confPath = conf
	}
	if state, ok := args["--state"].(string); ok {
		result.statePath = state
	}

	return result
}

func setupLogger(level logLevel) {
	switch level {
	case logVerbose:
		backsched.EnableDebug()
	case logNormal:
		break
	}
}

func handleCommand(opts cmdlineOpts) error {
	conf, err := backsched.ParseConfig(opts.confPath)
	if err != nil {
		return err
	}
	backsched.Debugf("config file:\n%v", conf.String())
	switch opts.command {
	case cmdNeedBackup:
		return backsched.NeedBackup(*conf, opts.statePath)
	}
	panic("command not found??")
}

func main() {
	opts := parseArgs()
	setupLogger(opts.logs)
	if err := handleCommand(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
