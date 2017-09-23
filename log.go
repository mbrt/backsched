package main

import (
	"fmt"
	"os"
)

var enableDebug = false

// Debugf prints the arguments if the debugging is enabled
func Debugf(format string, v ...interface{}) {
	if enableDebug {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}
