package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func PromptLine(in io.Reader, out io.Writer, label string, fallback string) (string, error) {
	if fallback != "" {
		fmt.Fprintf(out, "%s%s%s [%s]: ", bold, label, reset, fallback)
	} else {
		fmt.Fprintf(out, "%s%s%s: ", bold, label, reset)
	}
	reader, ok := in.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(in)
	}
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return value, nil
}

func PromptSecret(in io.Reader, out io.Writer, label string) (string, error) {
	return PromptLine(in, out, label, "")
}

func PromptYesNo(in io.Reader, out io.Writer, label string, fallback bool) (bool, error) {
	defaultLabel := "y"
	if !fallback {
		defaultLabel = "n"
	}
	value, err := PromptLine(in, out, label+" (y/n)", defaultLabel)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return fallback, nil
	}
}
