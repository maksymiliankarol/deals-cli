package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

const helpText = `usage:
  deals search <query>
  deals search -s <query>        run a search and save it
  deals search --saved <n>       run saved search #n
  deals saved                    list saved searches
  deals saved add <query>        save a search
  deals saved rm <n>             remove saved search #n
  deals buy --description <text> --email <email> --phone <phone>
  deals version                  print version
  deals help                     show this help`

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	var err error
	switch os.Args[1] {
	case "search":
		err = runSearch(os.Args[2:])
	case "buy":
		err = runBuy(os.Args[2:])
	case "saved":
		err = runSaved(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Println(version())
	case "help", "-h", "--help":
		fmt.Println(helpText)
	default:
		usage()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// usage is for wrong input: help goes to stderr and the exit code is 1.
func usage() {
	fmt.Fprintln(os.Stderr, helpText)
	os.Exit(1)
}

// set by release builds via -ldflags
var buildVersion string

func version() string {
	if buildVersion != "" {
		return "deals " + buildVersion
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return "deals " + info.Main.Version
	}
	return "deals (unknown version)"
}