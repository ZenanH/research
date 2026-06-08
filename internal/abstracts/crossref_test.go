package abstracts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrossrefAbstractCleansJATS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/works/10.1000%2Fexample" && r.URL.Path != "/works/10.1000/example" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"abstract":"<jats:p>This &amp; that abstract.</jats:p>"}}`))
	}))
	defer server.Close()

	client := &CrossrefClient{BaseURL: server.URL, HTTPClient: server.Client()}
	got, err := client.Abstract(context.Background(), "https://doi.org/10.1000/example")
	if err != nil {
		t.Fatal(err)
	}
	want := "This & that abstract."
	if got != want {
		t.Fatalf("Abstract() = %q, want %q", got, want)
	}
}
