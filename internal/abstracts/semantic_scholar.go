package abstracts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type SemanticScholarClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type semanticScholarPaperResponse struct {
	Abstract string `json:"abstract"`
}

func (c *SemanticScholarClient) Abstract(ctx context.Context, doi string) (string, error) {
	doi = normalizeDOI(doi)
	if doi == "" {
		return "", nil
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.semanticscholar.org/graph/v1"
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	endpoint := baseURL + "/paper/DOI:" + url.PathEscape(doi) + "?fields=abstract"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "research-cli (https://github.com/ZenanH/research)")
	if strings.TrimSpace(c.APIKey) != "" {
		req.Header.Set("x-api-key", strings.TrimSpace(c.APIKey))
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("semantic scholar request failed: %s", resp.Status)
	}
	var payload semanticScholarPaperResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.Abstract), nil
}
