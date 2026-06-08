package exporter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFilename(t *testing.T) {
	got := DefaultFilename("Computers and Geotechnics", "recent_100")
	want := "cag_recent_100.md"
	if got != want && got != "./"+want {
		t.Fatalf("DefaultFilename() = %q, want %q", got, want)
	}
}

func TestSuggestedFilename(t *testing.T) {
	if got := SuggestedFilename("Computers and Geotechnics", "recent", 100); got != "cag_100.md" {
		t.Fatalf("recent filename = %q", got)
	}
	if got := SuggestedFilename("Computers and Geotechnics", "keyword", 100); got != "cag_keywords_100.md" {
		t.Fatalf("keyword filename = %q", got)
	}
}

func TestResolveOutputPathUsesDirectory(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveOutputPath(dir, "Computers and Geotechnics", "recent", 100, ".")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "cag_100.md")
	if got != want {
		t.Fatalf("ResolveOutputPath() = %q, want %q", got, want)
	}
}

func TestResolveOutputPathUsesDefaultDirectory(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveOutputPath("", "Computers and Geotechnics", "keyword", 25, dir)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "cag_keywords_25.md")
	if got != want {
		t.Fatalf("ResolveOutputPath() = %q, want %q", got, want)
	}
}

func TestResolveOutputPathAvoidsOverwrite(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "cag_100.md")
	if err := os.WriteFile(existing, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveOutputPath(dir, "Computers and Geotechnics", "recent", 100, ".")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "cag_100_2.md")
	if got != want {
		t.Fatalf("ResolveOutputPath() = %q, want %q", got, want)
	}
}

func TestResolveOutputPathKeepsMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "custom.md")
	got, err := ResolveOutputPath(file, "Computers and Geotechnics", "recent", 100, ".")
	if err != nil {
		t.Fatal(err)
	}
	if got != file {
		t.Fatalf("ResolveOutputPath() = %q, want %q", got, file)
	}
}
