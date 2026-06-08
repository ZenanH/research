package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/ZenanH/research/internal/abstracts"
	"github.com/ZenanH/research/internal/exporter"
	"github.com/ZenanH/research/internal/openalex"
)

func RunJournal(ctx context.Context, opts Options, key string, semanticScholarKey string, defaultDir string) error {
	screen(opts.Out, "Recent Papers", "Export the latest OpenAlex journal articles to Markdown")
	journal, err := PromptLine(opts.In, opts.Out, "Journal name", "Computers and Geotechnics")
	if err != nil {
		return err
	}
	countText, err := PromptLine(opts.In, opts.Out, "Number of papers", "100")
	if err != nil {
		return err
	}
	count := parsePositiveInt(countText, 100)
	enrichAbstracts, err := PromptYesNo(opts.In, opts.Out, "Enrich missing abstracts?", true)
	if err != nil {
		return err
	}
	requireAbstract, err := PromptYesNo(opts.In, opts.Out, "Only include papers with abstracts?", false)
	if err != nil {
		return err
	}
	if enrichAbstracts {
		semanticScholarKey, err = ensureSemanticScholarKey(opts.In, opts.Out, semanticScholarKey)
		if err != nil {
			return err
		}
	}
	output, err := PromptLine(opts.In, opts.Out, "Output path", defaultOutput(journal, "recent", count, defaultDir))
	if err != nil {
		return err
	}
	output, err = exporter.ResolveOutputPath(output, journal, "recent", count, defaultDir)
	if err != nil {
		return err
	}

	client := openalex.NewClient(key)
	screen(opts.Out, "Recent Papers", "Resolving journal")
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
	screen(opts.Out, "Recent Papers", "Fetching papers")
	status(opts.Out, "Matched: "+source.DisplayName)
	status(opts.Out, "Fetching papers...")
	papers, err := client.SearchWorks(ctx, openalex.WorksQuery{
		SourceID:        source.ID,
		Count:           count,
		RequireAbstract: requireAbstract && !enrichAbstracts,
	})
	if err != nil {
		return err
	}
	if enrichAbstracts {
		status(opts.Out, "Enriching missing abstracts...")
		enricher := abstracts.NewEnricher(abstracts.Options{
			SemanticScholarKey: semanticScholarKey,
		})
		papers = enricher.Enrich(ctx, papers)
	}
	if requireAbstract {
		papers = abstracts.FilterWithAbstracts(papers)
	}
	status(opts.Out, "Writing Markdown...")
	if err := exporter.WriteCombined(output, exporter.ExportOptions{
		Title:          fmt.Sprintf("%s: Recent %d Papers", source.DisplayName, count),
		Journal:        journal,
		Source:         source,
		QueryType:      "recent",
		RequestedCount: count,
		GeneratedAt:    time.Now(),
	}, papers); err != nil {
		return err
	}
	summary := exporter.SummarizeAbstracts(papers)
	status(opts.Out, fmt.Sprintf("Done. Retrieved %d papers: %s", len(papers), output))
	status(opts.Out, fmt.Sprintf("Abstract coverage: %s (%d/%d)", summary.Coverage, summary.WithAbstracts, summary.Total))
	status(opts.Out, "Abstract sources: "+exporter.FormatAbstractSourceCounts(summary.SourceCounts))
	return nil
}
