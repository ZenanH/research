package openalex

import "testing"

func TestReconstructAbstract(t *testing.T) {
	index := map[string][]int{
		"OpenAlex": {0},
		"abstract": {2},
		"rebuilds": {1},
	}
	got := ReconstructAbstract(index)
	want := "OpenAlex rebuilds abstract"
	if got != want {
		t.Fatalf("ReconstructAbstract() = %q, want %q", got, want)
	}
}

func TestNormalizeOpenAlexID(t *testing.T) {
	got := normalizeOpenAlexID("https://openalex.org/S123456789")
	want := "S123456789"
	if got != want {
		t.Fatalf("normalizeOpenAlexID() = %q, want %q", got, want)
	}
}
