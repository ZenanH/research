package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ZenanH/research/internal/abstracts"
	"github.com/ZenanH/research/internal/config"
	"github.com/ZenanH/research/internal/exporter"
	"github.com/ZenanH/research/internal/model"
	"github.com/ZenanH/research/internal/openalex"
	"github.com/ZenanH/research/internal/tui"
)

type App struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Stdin   io.Reader
	Version string
}

func Run(args []string, version string) int {
	app := App{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
		Version: version,
	}
	if err := app.Run(context.Background(), args); err != nil {
		fmt.Fprintf(app.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func (a App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return tui.Run(ctx, tui.Options{
			In:      a.Stdin,
			Out:     a.Stdout,
			Err:     a.Stderr,
			Version: a.Version,
		})
	}

	switch args[0] {
	case "-h", "--help", "help":
		printHelp(a.Stdout)
		return nil
	case "-v", "--version", "version":
		fmt.Fprintf(a.Stdout, "research %s\n", a.Version)
		return nil
	case "journal":
		return a.runJournal(ctx, args[1:])
	case "search":
		return a.runSearch(ctx, args[1:])
	case "sources":
		return a.runSources(ctx, args[1:])
	case "config":
		return a.runConfig(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\nRun research --help for usage.", args[0])
	}
}

func (a App) runJournal(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("journal", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	name := fs.String("name", "", "journal name")
	count := fs.Int("count", 25, "number of papers")
	output := fs.String("output", "", "output Markdown path")
	openAlexKey := fs.String("openalex-key", "", "OpenAlex API key")
	requireAbstract := fs.Bool("require-abstract", false, "only export papers with abstracts from available sources")
	enrichAbstracts := fs.Bool("enrich-abstracts", true, "enrich missing abstracts with Crossref and Semantic Scholar")
	semanticScholarKey := fs.String("semantic-scholar-key", "", "Semantic Scholar API key")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*name) == "" {
		return errors.New("journal --name is required")
	}
	if *count <= 0 {
		return errors.New("journal --count must be greater than 0")
	}
	defaultDir := defaultOutputDir()
	resolvedOutput, err := exporter.ResolveOutputPath(*output, *name, "recent", *count, defaultDir)
	if err != nil {
		return err
	}
	*output = resolvedOutput
	key, err := a.requireKey(*openAlexKey)
	if err != nil {
		return err
	}
	client := openalex.NewClient(key)
	source, err := resolveSource(ctx, client, *name)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "Matched: %s\n", source.DisplayName)
	fmt.Fprintf(a.Stdout, "Fetching papers...\n")
	papers, err := client.SearchWorks(ctx, openalex.WorksQuery{
		SourceID:        source.ID,
		Count:           *count,
		RequireAbstract: *requireAbstract && !*enrichAbstracts,
	})
	if err != nil {
		return err
	}
	papers = a.enrichAndFilter(ctx, papers, *enrichAbstracts, *semanticScholarKey, *requireAbstract)
	return writeResult(*output, exporter.ExportOptions{
		Title:          fmt.Sprintf("%s: Recent %d Papers", source.DisplayName, *count),
		Journal:        *name,
		Source:         source,
		QueryType:      "recent",
		RequestedCount: *count,
		GeneratedAt:    time.Now(),
	}, papers, a.Stdout)
}

func (a App) runSearch(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	journal := fs.String("journal", "", "journal name")
	count := fs.Int("count", 25, "number of papers")
	keywords := fs.String("keywords", "", "comma-separated keywords")
	keywordMode := fs.String("keyword-mode", "any", "keyword mode: any or all")
	output := fs.String("output", "", "output Markdown path")
	openAlexKey := fs.String("openalex-key", "", "OpenAlex API key")
	requireAbstract := fs.Bool("require-abstract", false, "only export papers with abstracts from available sources")
	enrichAbstracts := fs.Bool("enrich-abstracts", true, "enrich missing abstracts with Crossref and Semantic Scholar")
	semanticScholarKey := fs.String("semantic-scholar-key", "", "Semantic Scholar API key")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*journal) == "" {
		return errors.New("search --journal is required")
	}
	if *count <= 0 {
		return errors.New("search --count must be greater than 0")
	}
	parsedKeywords := parseKeywords(*keywords)
	if len(parsedKeywords) == 0 {
		return errors.New("search --keywords is required")
	}
	if *keywordMode != "any" && *keywordMode != "all" {
		return errors.New("search --keyword-mode must be any or all")
	}
	defaultDir := defaultOutputDir()
	resolvedOutput, err := exporter.ResolveOutputPath(*output, *journal, "keyword", *count, defaultDir)
	if err != nil {
		return err
	}
	*output = resolvedOutput
	key, err := a.requireKey(*openAlexKey)
	if err != nil {
		return err
	}
	client := openalex.NewClient(key)
	source, err := resolveSource(ctx, client, *journal)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "Matched: %s\n", source.DisplayName)
	fmt.Fprintf(a.Stdout, "Searching papers...\n")
	papers, err := client.SearchWorks(ctx, openalex.WorksQuery{
		SourceID:        source.ID,
		Count:           *count,
		Keywords:        parsedKeywords,
		KeywordMode:     *keywordMode,
		RequireAbstract: *requireAbstract && !*enrichAbstracts,
	})
	if err != nil {
		return err
	}
	papers = a.enrichAndFilter(ctx, papers, *enrichAbstracts, *semanticScholarKey, *requireAbstract)
	return writeResult(*output, exporter.ExportOptions{
		Title:          fmt.Sprintf("%s: Keyword Search (%d Papers)", source.DisplayName, *count),
		Journal:        *journal,
		Source:         source,
		QueryType:      "keyword",
		RequestedCount: *count,
		Keywords:       parsedKeywords,
		KeywordMode:    *keywordMode,
		GeneratedAt:    time.Now(),
	}, papers, a.Stdout)
}

func (a App) enrichAndFilter(ctx context.Context, papers []model.Paper, enabled bool, semanticScholarKey string, requireAbstract bool) []model.Paper {
	if enabled {
		cfg, _, err := config.Load()
		if err == nil {
			semanticScholarKey = config.ResolveSemanticScholarKey(semanticScholarKey, cfg)
		} else {
			semanticScholarKey = firstNonEmpty(semanticScholarKey, os.Getenv(config.EnvSemanticScholarAPIKey))
		}
		fmt.Fprintf(a.Stdout, "Enriching missing abstracts...\n")
		enricher := abstracts.NewEnricher(abstracts.Options{
			SemanticScholarKey: semanticScholarKey,
		})
		papers = enricher.Enrich(ctx, papers)
	}
	if requireAbstract {
		papers = abstracts.FilterWithAbstracts(papers)
	}
	return papers
}

func (a App) runSources(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("sources", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	openAlexKey := fs.String("openalex-key", "", "OpenAlex API key")
	limit := fs.Int("limit", 5, "number of source candidates")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	query := strings.Join(fs.Args(), " ")
	if strings.TrimSpace(query) == "" {
		return errors.New("sources requires a journal query")
	}
	if *limit <= 0 {
		return errors.New("sources --limit must be greater than 0")
	}
	key, err := a.requireKey(*openAlexKey)
	if err != nil {
		return err
	}
	client := openalex.NewClient(key)
	sources, err := client.SearchSources(ctx, query, *limit)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		fmt.Fprintln(a.Stdout, "No sources found.")
		return nil
	}
	for i, source := range sources {
		fmt.Fprintf(a.Stdout, "%d. %s\n", i+1, source.DisplayName)
		fmt.Fprintf(a.Stdout, "   ISSN-L: %s\n", valueOrNA(source.ISSNL))
		fmt.Fprintf(a.Stdout, "   ISSN: %s\n", valueOrNA(strings.Join(source.ISSN, ", ")))
		fmt.Fprintf(a.Stdout, "   Works count: %d\n", source.WorksCount)
		fmt.Fprintf(a.Stdout, "   OpenAlex: %s\n", source.ID)
	}
	return nil
}

func (a App) runConfig(args []string) error {
	cfg, path, err := config.Load()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		fmt.Fprintf(a.Stdout, "Config path: %s\n", path)
		fmt.Fprintf(a.Stdout, "OpenAlex API key: %s\n", maskedKey(cfg.OpenAlexAPIKey))
		fmt.Fprintf(a.Stdout, "Semantic Scholar API key: %s\n", maskedKey(cfg.SemanticScholarAPIKey))
		fmt.Fprintf(a.Stdout, "Default output dir: %s\n", cfg.DefaultDir)
		fmt.Fprintf(a.Stdout, "Export mode: %s\n", cfg.ExportMode)
		return nil
	}
	switch args[0] {
	case "set":
		if len(args) < 2 {
			return errors.New("usage: research config set openalex-key|semantic-scholar-key")
		}
		switch args[1] {
		case "openalex-key":
			key, err := tui.PromptSecret(a.Stdin, a.Stdout, "Enter OpenAlex API key")
			if err != nil {
				return err
			}
			cfg.OpenAlexAPIKey = strings.TrimSpace(key)
			if cfg.OpenAlexAPIKey == "" {
				return errors.New("OpenAlex API key cannot be empty")
			}
			path, err := config.Save(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.Stdout, "Saved OpenAlex API key to %s\n", path)
			return nil
		case "semantic-scholar-key":
			key, err := tui.PromptSecret(a.Stdin, a.Stdout, "Enter Semantic Scholar API key")
			if err != nil {
				return err
			}
			cfg.SemanticScholarAPIKey = strings.TrimSpace(key)
			if cfg.SemanticScholarAPIKey == "" {
				return errors.New("Semantic Scholar API key cannot be empty")
			}
			path, err := config.Save(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.Stdout, "Saved Semantic Scholar API key to %s\n", path)
			return nil
		default:
			return errors.New("usage: research config set openalex-key|semantic-scholar-key")
		}
	case "unset":
		if len(args) < 2 {
			return errors.New("usage: research config unset openalex-key|semantic-scholar-key")
		}
		switch args[1] {
		case "openalex-key":
			cfg.OpenAlexAPIKey = ""
			path, err := config.Save(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.Stdout, "Removed OpenAlex API key from %s\n", path)
			return nil
		case "semantic-scholar-key":
			cfg.SemanticScholarAPIKey = ""
			path, err := config.Save(cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.Stdout, "Removed Semantic Scholar API key from %s\n", path)
			return nil
		default:
			return errors.New("usage: research config unset openalex-key|semantic-scholar-key")
		}
	default:
		return fmt.Errorf("unknown config command %q", args[0])
	}
}

func (a App) requireKey(cliKey string) (string, error) {
	cfg, _, err := config.Load()
	if err != nil {
		return "", err
	}
	key := config.ResolveOpenAlexKey(cliKey, cfg)
	if key == "" {
		return "", fmt.Errorf("OpenAlex API key required; pass --openalex-key, set %s, or run research config set openalex-key", config.EnvOpenAlexAPIKey)
	}
	return key, nil
}

func resolveSource(ctx context.Context, client *openalex.Client, query string) (model.Source, error) {
	sources, err := client.SearchSources(ctx, query, 5)
	if err != nil {
		return model.Source{}, err
	}
	if len(sources) == 0 {
		return model.Source{}, fmt.Errorf("no OpenAlex source found for %q", query)
	}
	return sources[0], nil
}

func writeResult(path string, opts exporter.ExportOptions, papers []model.Paper, out io.Writer) error {
	fmt.Fprintf(out, "Writing Markdown...\n")
	if err := exporter.WriteCombined(path, opts, papers); err != nil {
		return err
	}
	summary := exporter.SummarizeAbstracts(papers)
	fmt.Fprintf(out, "Done. Retrieved %d papers: %s\n", len(papers), path)
	fmt.Fprintf(out, "Abstract coverage: %s (%d/%d)\n", summary.Coverage, summary.WithAbstracts, summary.Total)
	fmt.Fprintf(out, "Abstract sources: %s\n", exporter.FormatAbstractSourceCounts(summary.SourceCounts))
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

func maskedKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "not set"
	}
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func valueOrNA(value string) string {
	if strings.TrimSpace(value) == "" {
		return "N/A"
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func defaultOutputDir() string {
	cfg, _, err := config.Load()
	if err != nil {
		return config.Default().DefaultDir
	}
	return cfg.DefaultDir
}

func printHelp(out io.Writer) {
	fmt.Fprint(out, `Research
OpenAlex journal paper exporter

Usage:
  research
  research journal --name "computers and geotechnics" --count 100 --output ./recent.md
  research search --journal "computers and geotechnics" --keywords "machine learning,DEM" --keyword-mode any --count 100 --output ./keywords.md
  research sources "computers and geotechnics"
  research config
  research config set openalex-key
  research config set semantic-scholar-key
  research config unset openalex-key
  research config unset semantic-scholar-key

Options:
  --openalex-key string  OpenAlex API key. If omitted, research checks OPENALEX_API_KEY and local config.
  --enrich-abstracts    Enrich missing abstracts with Crossref and Semantic Scholar. Enabled by default.
  --require-abstract    Only export papers with abstracts available from OpenAlex, Crossref, or Semantic Scholar.
  --semantic-scholar-key string  Semantic Scholar API key. If omitted, research checks SEMANTIC_SCHOLAR_API_KEY and local config, then uses anonymous access.
  --version             Show version.
`)
}
