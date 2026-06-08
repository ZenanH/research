package openalex

import (
	"context"
	"net/url"

	"github.com/ZenanH/research/internal/model"
)

type sourceResponse struct {
	Results []struct {
		ID          string   `json:"id"`
		DisplayName string   `json:"display_name"`
		ISSNL       string   `json:"issn_l"`
		ISSN        []string `json:"issn"`
		WorksCount  int      `json:"works_count"`
	} `json:"results"`
}

func (c *Client) SearchSources(ctx context.Context, query string, limit int) ([]model.Source, error) {
	if limit <= 0 {
		limit = 5
	}
	params := url.Values{}
	params.Set("search", query)
	params.Set("per-page", itoa(limit))

	var resp sourceResponse
	if err := c.get(ctx, "/sources", params, &resp); err != nil {
		return nil, err
	}
	sources := make([]model.Source, 0, len(resp.Results))
	for _, item := range resp.Results {
		sources = append(sources, model.Source{
			ID:          item.ID,
			DisplayName: item.DisplayName,
			ISSNL:       item.ISSNL,
			ISSN:        item.ISSN,
			WorksCount:  item.WorksCount,
		})
	}
	return sources, nil
}
