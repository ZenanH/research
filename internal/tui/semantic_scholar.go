package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/ZenanH/research/internal/config"
)

func ensureSemanticScholarKey(in io.Reader, out io.Writer, current string) (string, error) {
	current = strings.TrimSpace(current)
	if current != "" {
		return current, nil
	}

	screen(out, "Semantic Scholar", "Optional fallback for missing abstracts")
	fmt.Fprintln(out, "Semantic Scholar can be used without a key, but anonymous access has lower rate limits.")
	fmt.Fprintln(out)
	key, err := PromptSecret(in, out, "Enter Semantic Scholar API key (press Enter to use anonymous access)")
	if err != nil {
		return "", err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		note(out, "Using anonymous Semantic Scholar access for this run.")
		return "", nil
	}

	save, err := PromptYesNo(in, out, "Save this key for future runs?", true)
	if err != nil {
		return "", err
	}
	if save {
		cfg, path, err := config.Load()
		if err != nil {
			return "", err
		}
		cfg.SemanticScholarAPIKey = key
		if _, err := config.Save(cfg); err != nil {
			return "", err
		}
		status(out, "Saved Semantic Scholar key to "+path)
	}
	return key, nil
}
