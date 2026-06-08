package exporter

import (
	"strings"
	"testing"
	"time"

	"github.com/ZenanH/research/internal/model"
)

func TestCombinedMarkdownIncludesMetadataAndMissingAbstract(t *testing.T) {
	md := CombinedMarkdown(ExportOptions{
		Title:   "Computers and Geotechnics: Recent 1 Papers",
		Journal: "Computers and Geotechnics",
		Source: model.Source{
			ID:          "https://openalex.org/S123",
			DisplayName: "Computers and Geotechnics",
			ISSNL:       "0266-352X",
			ISSN:        []string{"0266-352X"},
		},
		QueryType:      "recent",
		RequestedCount: 1,
		GeneratedAt:    time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}, []model.Paper{
		{
			ID:              "https://openalex.org/W1",
			DisplayName:     "Paper without abstract",
			PublicationDate: "2026-01-01",
			Authors:         []string{"Ada Lovelace"},
		},
	})

	for _, want := range []string{
		"# Computers and Geotechnics: Recent 1 Papers",
		"- Source: OpenAlex",
		"- Retrieved count: 1",
		"_No abstract available from OpenAlex._",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}
}
