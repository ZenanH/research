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
		"- Papers with abstracts: 0",
		"- Abstract coverage: 0.0%",
		"- Abstract sources: OpenAlex 0, Crossref 0, Semantic Scholar 0, Missing 1",
		"- Abstract source: Missing",
		"_No abstract available from available sources._",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}
}
