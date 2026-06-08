package abstracts

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ZenanH/research/internal/model"
)

const EnvSemanticScholarAPIKey = "SEMANTIC_SCHOLAR_API_KEY"

type Enricher struct {
	Crossref        *CrossrefClient
	SemanticScholar *SemanticScholarClient
}

type Options struct {
	SemanticScholarKey string
	HTTPClient         *http.Client
}

func NewEnricher(opts Options) *Enricher {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Enricher{
		Crossref: &CrossrefClient{
			BaseURL:    "https://api.crossref.org",
			HTTPClient: httpClient,
		},
		SemanticScholar: &SemanticScholarClient{
			BaseURL:    "https://api.semanticscholar.org/graph/v1",
			APIKey:     opts.SemanticScholarKey,
			HTTPClient: httpClient,
		},
	}
}

func (e *Enricher) Enrich(ctx context.Context, papers []model.Paper) []model.Paper {
	for i := range papers {
		if strings.TrimSpace(papers[i].Abstract) != "" {
			continue
		}
		doi := normalizeDOI(papers[i].DOI)
		if doi == "" {
			continue
		}
		if e.Crossref != nil {
			if abstract, err := e.Crossref.Abstract(ctx, doi); err == nil && strings.TrimSpace(abstract) != "" {
				papers[i].Abstract = abstract
				papers[i].AbstractSource = model.AbstractSourceCrossref
				continue
			}
		}
		if e.SemanticScholar != nil && strings.TrimSpace(e.SemanticScholar.APIKey) != "" {
			if abstract, err := e.SemanticScholar.Abstract(ctx, doi); err == nil && strings.TrimSpace(abstract) != "" {
				papers[i].Abstract = abstract
				papers[i].AbstractSource = model.AbstractSourceSemanticScholar
				continue
			}
		}
		papers[i].AbstractSource = model.AbstractSourceMissing
	}
	return papers
}

func FilterWithAbstracts(papers []model.Paper) []model.Paper {
	filtered := make([]model.Paper, 0, len(papers))
	for _, paper := range papers {
		if strings.TrimSpace(paper.Abstract) != "" {
			filtered = append(filtered, paper)
		}
	}
	return filtered
}

func normalizeDOI(doi string) string {
	doi = strings.TrimSpace(doi)
	doi = strings.TrimPrefix(doi, "https://doi.org/")
	doi = strings.TrimPrefix(doi, "http://doi.org/")
	doi = strings.TrimPrefix(doi, "doi:")
	doi = strings.TrimPrefix(doi, "DOI:")
	return strings.TrimSpace(doi)
}
