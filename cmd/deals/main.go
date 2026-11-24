package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

const helpText = `usage:
  deals sell <query>             find buyers for what you sell
  deals sell -s <query>          run a search and save it
  deals sell --saved <n>         run saved search #n
  deals saved                    list saved searches
  deals saved add <query>        save a search
  deals saved rm <n>             remove saved search #n
  deals unlock <id>              get a payment link for listing #id
  deals unlocked                 list contacts you unlocked
  deals unlocked rm <id>         remove an unlocked contact
  deals buy                      post what you want to buy, so sellers can find you
      --description <text>       what you want (10-500 chars)
      --email <email>            your email
      --phone <phone>            your phone, e.g. +15551234567
  deals update                   update to the latest version
  deals version                  print version
  deals help                     show this help`

// title returns the bold app name, or "" when not on a colour terminal
// (same guard as logo()).
func title() string {
	if !isTerminal(os.Stdout) || os.Getenv("NO_COLOR") != "" {
		return ""
	}
	return "\n\x1b[1mM - Deals\x1b[m\n\n"
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	var notice <-chan string
	if os.Args[1] != "update" {
		notice = checkForUpdate()
	}

	var err error
	switch os.Args[1] {
	case "sell":
		err = runSearch(os.Args[2:])
	case "buy":
		err = runBuy(os.Args[2:])
	case "saved":
		err = runSaved(os.Args[2:])
	case "unlock":
		err = runUnlock(os.Args[2:])
	case "unlocked":
		err = runUnlocked(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Println(version())
	case "update":
		err = runUpdate(os.Args[2:])
	case "help", "-h", "--help":
		printHelp()
	default:
		usage()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if notice != nil {
		if n := <-notice; n != "" {
			fmt.Fprintln(os.Stderr, n)
		}
	}
}

// printHelp shows the logo (in a colour terminal) and the help text on stdout.
func printHelp() {
	fmt.Print(logo())
	fmt.Print(title())
	fmt.Println(helpText)
}

// usage is for wrong input: help goes to stderr and the exit code is 1.
func usage() {
	fmt.Print(title())
	fmt.Fprintln(os.Stderr, helpText)
	os.Exit(1)
}

// set by release builds via -ldflags
var buildVersion string

// currentVersion returns e.g. "0.1.0", or "" for local/dev builds.
func currentVersion() string {
	if buildVersion != "" {
		return buildVersion
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := strings.TrimPrefix(info.Main.Version, "v"); v != "" && v != "(devel)" {
			return v
		}
	}
	return ""
}

func version() string {
	if v := currentVersion(); v != "" {
		return "deals " + v
	}
	return "deals (unknown version)"
}
