package abstracts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSemanticScholarAbstractUsesAPIKey(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-api-key")
		_, _ = w.Write([]byte(`{"abstract":"Semantic Scholar abstract."}`))
	}))
	defer server.Close()

	client := &SemanticScholarClient{
		BaseURL:    server.URL,
		APIKey:     "secret",
		HTTPClient: server.Client(),
	}
	got, err := client.Abstract(context.Background(), "10.1000/example")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Semantic Scholar abstract." {
		t.Fatalf("Abstract() = %q", got)
	}
	if gotKey != "secret" {
		t.Fatalf("x-api-key = %q", gotKey)
	}
}
