package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var abbreviationTokens = regexp.MustCompile(`[a-zA-Z0-9]+`)

func DefaultFilename(journal string, suffix string) string {
	name := JournalAbbreviation(journal)
	if suffix != "" {
		name += "_" + suffix
	}
	return filepath.Join(".", name+".md")
}

func JournalAbbreviation(journal string) string {
	tokens := abbreviationTokens.FindAllString(journal, -1)
	var b strings.Builder
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		b.WriteByte(strings.ToLower(token)[0])
	}
	name := b.String()
	if name == "" {
		return "research"
	}
	return name
}

func SuggestedFilename(journal string, queryType string, count int) string {
	name := JournalAbbreviation(journal)
	if count <= 0 {
		count = 1
	}
	if queryType == "keyword" || queryType == "keywords" {
		return fmt.Sprintf("%s_keywords_%d.md", name, count)
	}
	return fmt.Sprintf("%s_%d.md", name, count)
}

func ResolveOutputPath(output string, journal string, queryType string, count int, defaultDir string) (string, error) {
	output = strings.TrimSpace(output)
	if strings.TrimSpace(defaultDir) == "" {
		defaultDir = "."
	}
	if output == "" {
		output = defaultDir
	}

	var path string
	if looksLikeDirectory(output) {
		path = filepath.Join(output, SuggestedFilename(journal, queryType, count))
	} else {
		path = output
	}
	path = strings.TrimSpace(path)
	if path == "" {
		path = filepath.Join(defaultDir, SuggestedFilename(journal, queryType, count))
	}
	if filepath.Ext(path) == "" {
		path += ".md"
	}
	return uniquePath(path), nil
}

func looksLikeDirectory(path string) bool {
	if path == "" {
		return true
	}
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, `\`) {
		return true
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return filepath.Ext(path) == ""
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); err != nil {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
