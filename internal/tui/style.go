package tui

import (
	"fmt"
	"io"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

func clear(out io.Writer) {
	fmt.Fprint(out, "\033[2J\033[H")
}

func screen(out io.Writer, title string, subtitle string) {
	clear(out)
	fmt.Fprintf(out, "%s%s%s\n", bold+cyan, title, reset)
	if subtitle != "" {
		fmt.Fprintf(out, "%s%s%s\n", dim, subtitle, reset)
	}
	fmt.Fprintln(out)
}

func status(out io.Writer, message string) {
	fmt.Fprintf(out, "%s%s%s\n", green, message, reset)
}

func note(out io.Writer, message string) {
	fmt.Fprintf(out, "%s%s%s\n", yellow, message, reset)
}
