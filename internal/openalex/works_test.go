package openalex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZenanH/research/internal/model"
)

func TestRecentWorksBuildsSourceFilterAndParsesPapers(t *testing.T) {
	var gotFilter string
	var gotSelect string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/works" {
			t.Fatalf("path = %s, want /works", r.URL.Path)
		}
		gotFilter = r.URL.Query().Get("filter")
		gotSelect = r.URL.Query().Get("select")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]string{"next_cursor": ""},
			"results": []map[string]any{
				{
					"id":               "https://openalex.org/W1",
					"doi":              "https://doi.org/10.1000/example",
					"display_name":     "A useful paper",
					"publication_date": "2026-01-02",
					"abstract_inverted_index": map[string][]int{
						"Useful":   {0},
						"abstract": {1},
					},
					"authorships": []map[string]any{
						{"author": map[string]string{"display_name": "Ada Lovelace"}},
					},
					"primary_location": map[string]string{
						"landing_page_url": "https://example.org/paper",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL
	papers, err := client.RecentWorks(context.Background(), "https://openalex.org/S123", 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotFilter != "primary_location.source.id:S123,type:article" {
		t.Fatalf("filter = %q", gotFilter)
	}
	if !strings.Contains(gotSelect, "abstract_inverted_index") {
		t.Fatalf("select = %q, want abstract_inverted_index", gotSelect)
	}
	if len(papers) != 1 {
		t.Fatalf("len(papers) = %d, want 1", len(papers))
	}
	if papers[0].Abstract != "Useful abstract" {
		t.Fatalf("abstract = %q", papers[0].Abstract)
	}
	if papers[0].AbstractSource != model.AbstractSourceOpenAlex {
		t.Fatalf("abstract source = %q", papers[0].AbstractSource)
	}
	if !strings.Contains(papers[0].Authors[0], "Ada") {
		t.Fatalf("authors = %#v", papers[0].Authors)
	}
}

func TestSearchWorksCanRequireAbstracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]string{"next_cursor": ""},
			"results": []map[string]any{
				{
					"id":               "https://openalex.org/W1",
					"display_name":     "Paper without abstract",
					"publication_date": "2026-01-01",
				},
				{
					"id":               "https://openalex.org/W2",
					"display_name":     "Paper with abstract",
					"publication_date": "2026-01-02",
					"abstract_inverted_index": map[string][]int{
						"Has":      {0},
						"abstract": {1},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL
	papers, err := client.SearchWorks(context.Background(), WorksQuery{
		SourceID:        "S123",
		Count:           2,
		RequireAbstract: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 1 {
		t.Fatalf("len(papers) = %d, want 1", len(papers))
	}
	if papers[0].ID != "https://openalex.org/W2" {
		t.Fatalf("paper ID = %q, want W2", papers[0].ID)
	}
}
