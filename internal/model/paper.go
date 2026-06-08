package model

type Paper struct {
	ID              string
	DOI             string
	DisplayName     string
	PublicationDate string
	Authors         []string
	Abstract        string
	AbstractSource  string
	LandingPageURL  string
}

const (
	AbstractSourceMissing         = "Missing"
	AbstractSourceOpenAlex        = "OpenAlex"
	AbstractSourceCrossref        = "Crossref"
	AbstractSourceSemanticScholar = "Semantic Scholar"
)
