package main

import (
	"fmt"
	"os"
	"strings"
)

// status shows msg on stderr while a slow step runs. Call the returned function
// when the step is done. Nothing is shown when stderr is not a terminal (pipes, scripts).
func status(msg string) func() {
	if fi, err := os.Stderr.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return func() {}
	}
	fmt.Fprint(os.Stderr, msg)
	return func() { fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", len(msg))+"\r") }
}