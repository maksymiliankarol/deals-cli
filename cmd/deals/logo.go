package main

import (
	"fmt"
	"os"
	"strings"
)

// Digits index logoColors; spaces are transparent.
var logoRows = [...]string{
	" 50 ",
	"4501",
	"4321",
	" 32 ",
}

var logoColors = [...]string{"255;0;0", "255;0;255", "0;0;255", "0;255;255", "0;255;0", "255;255;0"}

// logo renders two grid rows per terminal row using half blocks,
// with an empty line above and below. Returns "" if not on a colour terminal.
func logo() string {
	if !isTerminal(os.Stdout) || os.Getenv("NO_COLOR") != "" {
		return ""
	}
	c := func(ch byte) string { return logoColors[ch-'0'] }
	var b strings.Builder
	b.WriteByte('\n')
	for y := 0; y < len(logoRows); y += 2 {
		for x := range logoRows[y] {
			t, u := logoRows[y][x], logoRows[y+1][x]
			switch {
			case t == ' ' && u == ' ':
				b.WriteByte(' ')
			case u == ' ':
				fmt.Fprintf(&b, "\x1b[38;2;%sm▀\x1b[m", c(t))
			case t == ' ':
				fmt.Fprintf(&b, "\x1b[38;2;%sm▄\x1b[m", c(u))
			default:
				fmt.Fprintf(&b, "\x1b[38;2;%s;48;2;%sm▀\x1b[m", c(t), c(u))
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
