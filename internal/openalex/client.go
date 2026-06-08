package openalex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const DefaultBaseURL = "https://api.openalex.org"

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	if c.APIKey != "" {
		params.Set("api_key", c.APIKey)
	}
	endpoint := c.BaseURL + path
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "research-cli/dev")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var apiErr struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Message != "" {
			return fmt.Errorf("openalex request failed: %s: %s", resp.Status, apiErr.Message)
		}
		return fmt.Errorf("openalex request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
