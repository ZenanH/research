package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ZenanH/research/internal/exporter"
	"github.com/ZenanH/research/internal/openalex"
)

func RunSearch(ctx context.Context, opts Options, key string) error {
	fmt.Fprintln(opts.Out)
	fmt.Fprintln(opts.Out, "\033[1mKeyword search in journal\033[0m")
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
	output, err := PromptLine(opts.In, opts.Out, "Output path", defaultOutput(journal, "keywords"))
	if err != nil {
		return err
	}

	client := openalex.NewClient(key)
	fmt.Fprintln(opts.Out, "Resolving journal...")
	sources, err := client.SearchSources(ctx, journal, 5)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no OpenAlex source found for %q", journal)
	}
	sourceIndex := 0
	if len(sources) > 1 {
		fmt.Fprintln(opts.Out, "Source candidates:")
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
	fmt.Fprintf(opts.Out, "Matched: %s\n", source.DisplayName)
	fmt.Fprintf(opts.Out, "Searching papers...\n")
	papers, err := client.SearchWorks(ctx, openalex.WorksQuery{
		SourceID:    source.ID,
		Count:       count,
		Keywords:    keywords,
		KeywordMode: mode,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(opts.Out, "Writing Markdown...\n")
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
	fmt.Fprintf(opts.Out, "Done. Retrieved %d papers: %s\n", len(papers), output)
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
