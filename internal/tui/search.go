package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ZenanH/research/internal/abstracts"
	"github.com/ZenanH/research/internal/exporter"
	"github.com/ZenanH/research/internal/openalex"
)

func RunSearch(ctx context.Context, opts Options, key string) error {
	screen(opts.Out, "Keyword Search", "Export recent journal articles matching title or abstract keywords")
	journal, err := PromptLine(opts.In, opts.Out, "Journal name", "Computers and Geotechnics")
	if err != nil {
		return err
	}
	countText, err := PromptLine(opts.In, opts.Out, "Number of papers", "100")
	if err != nil {
		return err
	}
	count := parsePositiveInt(countText, 100)
	keywordsText, err := PromptLine(opts.In, opts.Out, "Keywords", "machine learning, DEM, slope stability")
	if err != nil {
		return err
	}
	keywords := parseKeywords(keywordsText)
	if len(keywords) == 0 {
		return fmt.Errorf("keywords are required")
	}
	mode, err := PromptLine(opts.In, opts.Out, "Keyword mode (any/all)", "any")
	if err != nil {
		return err
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "any" && mode != "all" {
		mode = "any"
	}
	enrichAbstracts, err := PromptYesNo(opts.In, opts.Out, "Enrich missing abstracts?", true)
	if err != nil {
		return err
	}
	requireAbstract, err := PromptYesNo(opts.In, opts.Out, "Only include papers with abstracts?", false)
	if err != nil {
		return err
	}
	output, err := PromptLine(opts.In, opts.Out, "Output path", defaultOutput(journal, "keywords"))
	if err != nil {
		return err
	}

	client := openalex.NewClient(key)
	screen(opts.Out, "Keyword Search", "Resolving journal")
	status(opts.Out, "Resolving journal...")
	sources, err := client.SearchSources(ctx, journal, 5)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no OpenAlex source found for %q", journal)
	}
	sourceIndex := 0
	if len(sources) > 1 {
		screen(opts.Out, "Choose Source", "Multiple OpenAlex source candidates were found")
		for i, source := range sources {
			fmt.Fprintf(opts.Out, "  %d. %s | ISSN-L: %s | Works: %d | %s\n", i+1, source.DisplayName, valueOrNA(source.ISSNL), source.WorksCount, source.ID)
		}
		choice, err := PromptLine(opts.In, opts.Out, "Choose source", "1")
		if err != nil {
			return err
		}
		sourceIndex = parsePositiveInt(choice, 1) - 1
		if sourceIndex < 0 || sourceIndex >= len(sources) {
			sourceIndex = 0
		}
	}
	source := sources[sourceIndex]
	screen(opts.Out, "Keyword Search", "Searching papers")
	status(opts.Out, "Matched: "+source.DisplayName)
	status(opts.Out, "Searching papers...")
	papers, err := client.SearchWorks(ctx, openalex.WorksQuery{
		SourceID:        source.ID,
		Count:           count,
		Keywords:        keywords,
		KeywordMode:     mode,
		RequireAbstract: requireAbstract && !enrichAbstracts,
	})
	if err != nil {
		return err
	}
	if enrichAbstracts {
		status(opts.Out, "Enriching missing abstracts...")
		enricher := abstracts.NewEnricher(abstracts.Options{
			SemanticScholarKey: os.Getenv(abstracts.EnvSemanticScholarAPIKey),
		})
		papers = enricher.Enrich(ctx, papers)
	}
	if requireAbstract {
		papers = abstracts.FilterWithAbstracts(papers)
	}
	status(opts.Out, "Writing Markdown...")
	if err := exporter.WriteCombined(output, exporter.ExportOptions{
		Title:          fmt.Sprintf("%s: Keyword Search (%d Papers)", source.DisplayName, count),
		Journal:        journal,
		Source:         source,
		QueryType:      "keyword",
		RequestedCount: count,
		Keywords:       keywords,
		KeywordMode:    mode,
		GeneratedAt:    time.Now(),
	}, papers); err != nil {
		return err
	}
	status(opts.Out, fmt.Sprintf("Done. Retrieved %d papers: %s", len(papers), output))
	return nil
}

func parseKeywords(input string) []string {
	parts := strings.Split(input, ",")
	keywords := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			keywords = append(keywords, part)
		}
	}
	return keywords
}
