package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	EnvOpenAlexAPIKey = "OPENALEX_API_KEY"
)

type Config struct {
	OpenAlexAPIKey string
	DefaultDir     string
	ExportMode     string
}

func Default() Config {
	return Config{
		DefaultDir: "./research-outputs",
		ExportMode: "combined",
	}
}

func Load() (Config, string, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, "", err
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, path, nil
	}
	if err != nil {
		return cfg, path, err
	}
	defer file.Close()

	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"`)

		switch section + "." + key {
		case "openalex.api_key":
			cfg.OpenAlexAPIKey = value
		case "output.default_dir":
			cfg.DefaultDir = value
		case "export.mode":
			cfg.ExportMode = value
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, path, err
	}
	return cfg, path, nil
}

func Save(cfg Config) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return path, err
	}
	content := fmt.Sprintf(`[openalex]
api_key = "%s"

[output]
default_dir = "%s"

[export]
mode = "%s"
`, escapeTOML(cfg.OpenAlexAPIKey), escapeTOML(cfg.DefaultDir), escapeTOML(cfg.ExportMode))
	return path, os.WriteFile(path, []byte(content), 0o600)
}

func ResolveOpenAlexKey(cliKey string, cfg Config) string {
	if strings.TrimSpace(cliKey) != "" {
		return strings.TrimSpace(cliKey)
	}
	if env := strings.TrimSpace(os.Getenv(EnvOpenAlexAPIKey)); env != "" {
		return env
	}
	return strings.TrimSpace(cfg.OpenAlexAPIKey)
}

func Path() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "research", "config.toml"), nil
	case "windows":
		base := os.Getenv("AppData")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, "research", "config.toml"), nil
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "research", "config.toml"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "research", "config.toml"), nil
	}
}

func escapeTOML(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}
