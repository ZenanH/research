package exporter

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZenanH/research/internal/model"
)

type ExportOptions struct {
	Title          string
	Journal        string
	Source         model.Source
	QueryType      string
	RequestedCount int
	Keywords       []string
	KeywordMode    string
	GeneratedAt    time.Time
}

func WriteCombined(path string, opts ExportOptions, papers []model.Paper) error {
	content := CombinedMarkdown(opts, papers)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func CombinedMarkdown(opts ExportOptions, papers []model.Paper) string {
	if opts.GeneratedAt.IsZero() {
		opts.GeneratedAt = time.Now()
	}
	if opts.Title == "" {
		opts.Title = fmt.Sprintf("%s: Recent %d Papers", opts.Source.DisplayName, opts.RequestedCount)
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s\n\n", escapeMD(opts.Title))
	buf.WriteString("## Metadata\n\n")
	fmt.Fprintf(&buf, "- Journal: %s\n", escapeMD(firstNonEmpty(opts.Journal, opts.Source.DisplayName)))
	buf.WriteString("- Source: OpenAlex\n")
	fmt.Fprintf(&buf, "- Source ID: %s\n", escapeMD(opts.Source.ID))
	fmt.Fprintf(&buf, "- ISSN-L: %s\n", escapeMD(emptyAsNA(opts.Source.ISSNL)))
	fmt.Fprintf(&buf, "- ISSN: %s\n", escapeMD(emptyAsNA(strings.Join(opts.Source.ISSN, ", "))))
	fmt.Fprintf(&buf, "- Requested count: %d\n", opts.RequestedCount)
	fmt.Fprintf(&buf, "- Retrieved count: %d\n", len(papers))
	abstractCount := papersWithAbstracts(papers)
	fmt.Fprintf(&buf, "- Papers with abstracts: %d\n", abstractCount)
	fmt.Fprintf(&buf, "- Abstract coverage: %s\n", percent(abstractCount, len(papers)))
	fmt.Fprintf(&buf, "- Abstract sources: %s\n", escapeMD(formatSourceCounts(abstractSourceCounts(papers))))
	fmt.Fprintf(&buf, "- Query type: %s\n", escapeMD(opts.QueryType))
	buf.WriteString("- Sort: publication_date:desc\n")
	fmt.Fprintf(&buf, "- Generated at: %s\n\n", opts.GeneratedAt.Format(time.RFC3339))

	if len(opts.Keywords) > 0 {
		buf.WriteString("## Query\n\n")
		fmt.Fprintf(&buf, "- Keywords: %s\n", escapeMD(strings.Join(opts.Keywords, ", ")))
		fmt.Fprintf(&buf, "- Keyword mode: %s\n\n", escapeMD(firstNonEmpty(opts.KeywordMode, "any")))
	}

	buf.WriteString("## Index\n\n")
	buf.WriteString("| # | Date | Title | Authors |\n")
	buf.WriteString("|---:|---|---|---|\n")
	for i, paper := range papers {
		fmt.Fprintf(&buf, "| %d | %s | %s | %s |\n",
			i+1,
			escapeTable(paper.PublicationDate),
			escapeTable(paper.DisplayName),
			escapeTable(authorsSummary(paper.Authors)),
		)
	}

	buf.WriteString("\n## Papers\n\n")
	for i, paper := range papers {
		fmt.Fprintf(&buf, "### %d. %s\n\n", i+1, escapeMD(paper.DisplayName))
		fmt.Fprintf(&buf, "- Date: %s\n", escapeMD(emptyAsNA(paper.PublicationDate)))
		fmt.Fprintf(&buf, "- Authors: %s\n", escapeMD(emptyAsNA(strings.Join(paper.Authors, ", "))))
		fmt.Fprintf(&buf, "- DOI: %s\n", escapeMD(emptyAsNA(paper.DOI)))
		fmt.Fprintf(&buf, "- OpenAlex: %s\n", escapeMD(emptyAsNA(paper.ID)))
		fmt.Fprintf(&buf, "- Abstract source: %s\n", escapeMD(emptyAsNA(abstractSource(paper))))
		fmt.Fprintf(&buf, "- Publisher page: %s\n\n", escapeMD(emptyAsNA(paper.LandingPageURL)))
		buf.WriteString("#### Abstract\n\n")
		if strings.TrimSpace(paper.Abstract) == "" {
			buf.WriteString("_No abstract available from configured sources._\n\n")
		} else {
			fmt.Fprintf(&buf, "%s\n\n", strings.TrimSpace(paper.Abstract))
		}
	}
	return buf.String()
}

func abstractSource(paper model.Paper) string {
	if strings.TrimSpace(paper.AbstractSource) != "" {
		return paper.AbstractSource
	}
	if strings.TrimSpace(paper.Abstract) != "" {
		return model.AbstractSourceOpenAlex
	}
	return model.AbstractSourceMissing
}

func papersWithAbstracts(papers []model.Paper) int {
	count := 0
	for _, paper := range papers {
		if strings.TrimSpace(paper.Abstract) != "" {
			count++
		}
	}
	return count
}

func abstractSourceCounts(papers []model.Paper) map[string]int {
	counts := map[string]int{
		model.AbstractSourceOpenAlex:        0,
		model.AbstractSourceCrossref:        0,
		model.AbstractSourceSemanticScholar: 0,
		model.AbstractSourceMissing:         0,
	}
	for _, paper := range papers {
		counts[abstractSource(paper)]++
	}
	return counts
}

func formatSourceCounts(counts map[string]int) string {
	parts := []string{
		fmt.Sprintf("OpenAlex %d", counts[model.AbstractSourceOpenAlex]),
		fmt.Sprintf("Crossref %d", counts[model.AbstractSourceCrossref]),
		fmt.Sprintf("Semantic Scholar %d", counts[model.AbstractSourceSemanticScholar]),
		fmt.Sprintf("Missing %d", counts[model.AbstractSourceMissing]),
	}
	return strings.Join(parts, ", ")
}

func percent(n int, total int) string {
	if total == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.1f%%", float64(n)*100/float64(total))
}

func authorsSummary(authors []string) string {
	switch len(authors) {
	case 0:
		return "N/A"
	case 1, 2, 3:
		return strings.Join(authors, ", ")
	default:
		return strings.Join(authors[:3], ", ") + " et al."
	}
}

func emptyAsNA(s string) string {
	if strings.TrimSpace(s) == "" {
		return "N/A"
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func escapeTable(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(escapeMD(s), "|", `\|`)
}

func escapeMD(s string) string {
	return html.EscapeString(s)
}
