package main

import (
	"fmt"
	"os"

	"github.com/gen2brain/dlgs"

	"github.com/mbrt/backsched/internal/errors"
)

func main() {
	yes, err := dlgs.Question("Question", "Are you sure you want to format this media?", true)
	if err != nil {
		panic(err)
	}
	println(yes)
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
}

func test() {
	// TEST
	passwd, _, err := dlgs.Password("Password", "Enter your API key:")
	if err != nil {
		panic(err)
	}
	fmt.Println(passwd)
}

func fatal(err error) {
	stderrPrintf("Error: %v\n", err)
	if det := errors.Details(err); det != "" {
		stderrPrintf("\nNote: %s\n", det)
	}
	os.Exit(1)
}

func stderrPrintf(format string, a ...interface{}) {
	/* #nosec */
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
