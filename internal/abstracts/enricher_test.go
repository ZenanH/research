package abstracts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZenanH/research/internal/model"
)

func TestEnricherUsesCrossrefForMissingAbstracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":{"abstract":"<p>Crossref abstract.</p>"}}`))
	}))
	defer server.Close()

	enricher := &Enricher{
		Crossref: &CrossrefClient{BaseURL: server.URL, HTTPClient: server.Client()},
	}
	papers := enricher.Enrich(context.Background(), []model.Paper{
		{DOI: "10.1000/example", AbstractSource: model.AbstractSourceMissing},
	})
	if papers[0].Abstract != "Crossref abstract." {
		t.Fatalf("abstract = %q", papers[0].Abstract)
	}
	if papers[0].AbstractSource != model.AbstractSourceCrossref {
		t.Fatalf("source = %q", papers[0].AbstractSource)
	}
}

func TestFilterWithAbstracts(t *testing.T) {
	papers := FilterWithAbstracts([]model.Paper{
		{DisplayName: "missing"},
		{DisplayName: "present", Abstract: "Abstract."},
	})
	if len(papers) != 1 || papers[0].DisplayName != "present" {
		t.Fatalf("filtered papers = %#v", papers)
	}
}
