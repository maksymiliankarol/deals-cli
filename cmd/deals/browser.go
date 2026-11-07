package main

import (
	"os/exec"
	"runtime"
)

// openBrowser tries to open url in the default browser. Failure is not fatal:
// the caller still prints the link.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}