package config

import "testing"

func TestSaveLoadSemanticScholarKey(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("AppData", "")

	_, err := Save(Config{
		OpenAlexAPIKey:        "openalex",
		SemanticScholarAPIKey: "semantic",
		DefaultDir:            "./out",
		ExportMode:            "combined",
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SemanticScholarAPIKey != "semantic" {
		t.Fatalf("SemanticScholarAPIKey = %q", cfg.SemanticScholarAPIKey)
	}
}

func TestResolveSemanticScholarKeyOrder(t *testing.T) {
	t.Setenv(EnvSemanticScholarAPIKey, "")
	cfg := Config{SemanticScholarAPIKey: "config"}
	if got := ResolveSemanticScholarKey("cli", cfg); got != "cli" {
		t.Fatalf("cli key = %q", got)
	}
	t.Setenv(EnvSemanticScholarAPIKey, "env")
	if got := ResolveSemanticScholarKey("", cfg); got != "env" {
		t.Fatalf("env key = %q", got)
	}
	t.Setenv(EnvSemanticScholarAPIKey, "")
	if got := ResolveSemanticScholarKey("", cfg); got != "config" {
		t.Fatalf("config key = %q", got)
	}
}
