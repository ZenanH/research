package openalex

import (
	"context"
	"net/url"
	"strings"

	"github.com/ZenanH/research/internal/model"
)

type worksResponse struct {
	Meta struct {
		NextCursor string `json:"next_cursor"`
	} `json:"meta"`
	Results []workResult `json:"results"`
}

type workResult struct {
	ID                    string            `json:"id"`
	DOI                   string            `json:"doi"`
	DisplayName           string            `json:"display_name"`
	PublicationDate       string            `json:"publication_date"`
	AbstractInvertedIndex map[string][]int  `json:"abstract_inverted_index"`
	Authorships           []authorship      `json:"authorships"`
	PrimaryLocation       primaryLocation   `json:"primary_location"`
	Locations             []primaryLocation `json:"locations"`
}

type authorship struct {
	Author struct {
		DisplayName string `json:"display_name"`
	} `json:"author"`
}

type primaryLocation struct {
	LandingPageURL string `json:"landing_page_url"`
}

type WorksQuery struct {
	SourceID    string
	Count       int
	Keywords    []string
	KeywordMode string
}

func (c *Client) RecentWorks(ctx context.Context, sourceID string, count int) ([]model.Paper, error) {
	return c.fetchWorks(ctx, WorksQuery{
		SourceID: sourceID,
		Count:    count,
	})
}

func (c *Client) SearchWorks(ctx context.Context, query WorksQuery) ([]model.Paper, error) {
	return c.fetchWorks(ctx, query)
}

func (c *Client) fetchWorks(ctx context.Context, query WorksQuery) ([]model.Paper, error) {
	if query.Count <= 0 {
		query.Count = 25
	}
	if query.KeywordMode == "" {
		query.KeywordMode = "any"
	}
	cursor := "*"
	papers := make([]model.Paper, 0, query.Count)
	perPage := 100

	for len(papers) < query.Count && cursor != "" {
		params := url.Values{}
		params.Set("filter", "primary_location.source.id:"+normalizeOpenAlexID(query.SourceID)+",type:article")
		params.Set("sort", "publication_date:desc")
		params.Set("per-page", itoa(perPage))
		params.Set("cursor", cursor)
		if len(query.Keywords) > 0 {
			params.Set("search", strings.Join(query.Keywords, " "))
		}

		var resp worksResponse
		if err := c.get(ctx, "/works", params, &resp); err != nil {
			return nil, err
		}
		for _, result := range resp.Results {
			paper := result.toPaper()
			if len(query.Keywords) == 0 || matchesKeywords(paper, query.Keywords, query.KeywordMode) {
				papers = append(papers, paper)
				if len(papers) == query.Count {
					break
				}
			}
		}
		if resp.Meta.NextCursor == cursor {
			break
		}
		cursor = resp.Meta.NextCursor
		if len(resp.Results) == 0 {
			break
		}
	}
	return papers, nil
}

func normalizeOpenAlexID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.TrimPrefix(id, "https://openalex.org/")
	id = strings.TrimPrefix(id, "http://openalex.org/")
	return id
}

func (w workResult) toPaper() model.Paper {
	authors := make([]string, 0, len(w.Authorships))
	seen := make(map[string]bool)
	for _, authorship := range w.Authorships {
		name := strings.TrimSpace(authorship.Author.DisplayName)
		if name != "" && !seen[name] {
			authors = append(authors, name)
			seen[name] = true
		}
	}
	return model.Paper{
		ID:              w.ID,
		DOI:             w.DOI,
		DisplayName:     w.DisplayName,
		PublicationDate: w.PublicationDate,
		Authors:         authors,
		Abstract:        ReconstructAbstract(w.AbstractInvertedIndex),
		LandingPageURL:  firstNonEmpty(w.PrimaryLocation.LandingPageURL, firstLocationURL(w.Locations)),
	}
}

func matchesKeywords(paper model.Paper, keywords []string, mode string) bool {
	haystack := strings.ToLower(paper.DisplayName + "\n" + paper.Abstract)
	matches := 0
	for _, keyword := range keywords {
		keyword = strings.ToLower(strings.TrimSpace(keyword))
		if keyword == "" {
			continue
		}
		if strings.Contains(haystack, keyword) {
			matches++
		} else if mode == "all" {
			return false
		}
	}
	if mode == "all" {
		return matches > 0
	}
	return matches > 0
}

func firstLocationURL(locations []primaryLocation) string {
	for _, location := range locations {
		if location.LandingPageURL != "" {
			return location.LandingPageURL
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
