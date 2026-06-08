package tui

import (
	"strconv"
	"strings"
)

func parsePositiveInt(input string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
