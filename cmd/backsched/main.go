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

	defaultConfFile = "~/.backsched/config.yaml"
)

type cmdlineOpts struct {
	command    commandType
	logs       logLevel
	confPath   string
	notifySend bool
}

func parseArgs() cmdlineOpts {
	usage := `backsched.

Usage:
  backsched need-backup [options] [-n]
  backsched backup [options]

Options:
  -h --help               Show this screen.
  --verbose               Verbose output.
  --config=<conf-file>    Use the specified config file.
  -n --notify-send        Use notify-send for GUI feedback.`

	args, _ := docopt.Parse(usage, nil, true, "Backup Scheduler 0.1", false)

	result := cmdlineOpts{
		cmdDoBackup,
		logNormal,
		defaultConfFile,
		false,
	}

	if args["need-backup"].(bool) {
		result.command = cmdNeedBackup
	}
	if args["--verbose"].(bool) {
		result.logs = logVerbose
	}
	if conf, ok := args["--config"].(string); ok {
		result.confPath = conf
	}
	result.notifySend = args["--notify-send"].(bool)

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
	state, err := backsched.LoadState(conf.StatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading saved state: %v\n", err)
		state = &backsched.State{}
	}

	backsched.Debugf("config file:\n%v", conf.String())
	switch opts.command {
	case cmdNeedBackup:
		return backsched.NeedBackup(*conf, *state, opts.notifySend)
	case cmdDoBackup:
		return backsched.Backup(*conf, *state)
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
