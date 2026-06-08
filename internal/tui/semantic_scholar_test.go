package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestEnsureSemanticScholarKeyUsesAnonymousWhenSkipped(t *testing.T) {
	var out bytes.Buffer
	got, err := ensureSemanticScholarKey(strings.NewReader("\n"), &out, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("key = %q, want empty anonymous key", got)
	}
	output := out.String()
	for _, want := range []string{
		"Semantic Scholar",
		"press Enter to use anonymous access",
		"Using anonymous Semantic Scholar access",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestEnsureSemanticScholarKeyReturnsConfiguredKeyWithoutPrompt(t *testing.T) {
	var out bytes.Buffer
	got, err := ensureSemanticScholarKey(strings.NewReader("ignored\n"), &out, "configured")
	if err != nil {
		t.Fatal(err)
	}
	if got != "configured" {
		t.Fatalf("key = %q", got)
	}
	if out.Len() != 0 {
		t.Fatalf("unexpected prompt output: %q", out.String())
	}
}
