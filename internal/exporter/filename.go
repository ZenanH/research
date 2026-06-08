package exporter

import (
	"path/filepath"
	"regexp"
	"strings"
)

var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func DefaultFilename(journal string, suffix string) string {
	name := strings.ToLower(strings.TrimSpace(journal))
	name = unsafeFilenameChars.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = "research"
	}
	if suffix != "" {
		name += "_" + suffix
	}
	return filepath.Join(".", name+".md")
}
