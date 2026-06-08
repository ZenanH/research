package abstracts

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type CrossrefClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type crossrefWorkResponse struct {
	Message struct {
		Abstract string `json:"abstract"`
	} `json:"message"`
}

var tagPattern = regexp.MustCompile(`<[^>]+>`)
var whitespacePattern = regexp.MustCompile(`\s+`)

func (c *CrossrefClient) Abstract(ctx context.Context, doi string) (string, error) {
	doi = normalizeDOI(doi)
	if doi == "" {
		return "", nil
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.crossref.org"
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	endpoint := baseURL + "/works/" + url.PathEscape(doi)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "research-cli (https://github.com/ZenanH/research)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("crossref request failed: %s", resp.Status)
	}
	var payload crossrefWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return cleanAbstract(payload.Message.Abstract), nil
}

func cleanAbstract(input string) string {
	input = html.UnescapeString(input)
	input = tagPattern.ReplaceAllString(input, " ")
	input = whitespacePattern.ReplaceAllString(input, " ")
	return strings.TrimSpace(input)
}
