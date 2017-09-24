package backsched

import (
	"fmt"
	"os"
)

var enableDebug = false

// EnableDebug enables debug logging
func EnableDebug() {
	enableDebug = true
}

// Debugf prints the arguments if the debugging is enabled
func Debugf(format string, v ...interface{}) {
	if enableDebug {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}
