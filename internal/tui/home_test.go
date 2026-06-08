package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunPromptsForOpenAlexKeyBeforeMenu(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENALEX_API_KEY", "")

	var out bytes.Buffer
	in := strings.NewReader("test-openalex-key\nn\n4\n")

	err := Run(context.Background(), Options{
		In:  in,
		Out: &out,
		Err: &out,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{
		"OpenAlex API key required",
		"Enter OpenAlex API key",
		"Save this key for future runs?",
		"Choose a workflow",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}
