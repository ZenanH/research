package exporter

import "testing"

func TestDefaultFilename(t *testing.T) {
	got := DefaultFilename("Computers and Geotechnics", "recent_100")
	want := "computers_and_geotechnics_recent_100.md"
	if got != want && got != "./"+want {
		t.Fatalf("DefaultFilename() = %q, want %q", got, want)
	}
}
