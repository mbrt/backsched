package main

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
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
	command  commandType
	logs     logLevel
	confPath string
}

func parseArgs() cmdlineOpts {
	usage := `backsched.

Usage:
  backsched need-backup [options]
  backsched backup [options]

Options:
  -h --help               Show this screen.
  --verbose               Verbose output.
  --config=<conf-file>    Use the specified config file.`

	args, _ := docopt.Parse(usage, nil, true, "Backup Scheduler 0.1", true)

	result := cmdlineOpts{
		cmdDoBackup,
		logNormal,
		defaultConfFile,
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

	return result
}

func setupLogger(level logLevel) {
	switch level {
	case logVerbose:
		enableDebug = true
	case logNormal:
		break
	}
}

func handleCommand(opts cmdlineOpts) error {
	conf, err := ParseConfig(opts.confPath)
	if err != nil {
		return err
	}
	Debugf("config: %v\n", conf.String())
	return nil
}

func main() {
	opts := parseArgs()
	setupLogger(opts.logs)
	if err := handleCommand(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
