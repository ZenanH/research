package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ZenanH/research/internal/config"
	"github.com/ZenanH/research/internal/exporter"
)

type Options struct {
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Version string
}

func Run(ctx context.Context, opts Options) error {
	if _, ok := opts.In.(*bufio.Reader); !ok {
		opts.In = bufio.NewReader(opts.In)
	}
	cfg, path, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Fprintln(opts.Out, "\033[1mResearch\033[0m")
	fmt.Fprintln(opts.Out, "OpenAlex journal paper exporter")
	fmt.Fprintln(opts.Out)

	key := config.ResolveOpenAlexKey("", cfg)
	if key == "" {
		fmt.Fprintln(opts.Out, "OpenAlex API key required.")
		fmt.Fprintln(opts.Out, "Get a free key at https://openalex.org/settings/api")
		key, err = PromptSecret(opts.In, opts.Out, "Enter OpenAlex API key")
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("OpenAlex API key is required")
		}
		save, err := PromptYesNo(opts.In, opts.Out, "Save this key for future runs?", true)
		if err != nil {
			return err
		}
		if save {
			cfg.OpenAlexAPIKey = key
			if _, err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(opts.Out, "Saved key to %s\n\n", path)
		}
	}

	for {
		fmt.Fprintln(opts.Out, "Choose a workflow")
		fmt.Fprintln(opts.Out, "  1. Recent papers from journal")
		fmt.Fprintln(opts.Out, "  2. Keyword search in journal")
		fmt.Fprintln(opts.Out, "  3. Settings")
		fmt.Fprintln(opts.Out, "  4. Exit")
		choice, err := PromptLine(opts.In, opts.Out, "Selection", "1")
		if err != nil {
			return err
		}
		switch strings.TrimSpace(choice) {
		case "1":
			return RunJournal(ctx, opts, key)
		case "2":
			return RunSearch(ctx, opts, key)
		case "3":
			fmt.Fprintf(opts.Out, "Config path: %s\n", path)
			fmt.Fprintf(opts.Out, "OpenAlex API key: %s\n", maskedKey(cfg.OpenAlexAPIKey))
			fmt.Fprintf(opts.Out, "Default output dir: %s\n", cfg.DefaultDir)
			fmt.Fprintf(opts.Out, "Export mode: %s\n\n", cfg.ExportMode)
		case "4", "q", "quit", "exit":
			return nil
		default:
			fmt.Fprintln(opts.Out, "Please choose 1, 2, 3, or 4.")
		}
	}
}

func defaultOutput(journal string, suffix string) string {
	return exporter.DefaultFilename(journal, suffix)
}

func maskedKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "not set"
	}
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func valueOrNA(value string) string {
	if strings.TrimSpace(value) == "" {
		return "N/A"
	}
	return value
}
