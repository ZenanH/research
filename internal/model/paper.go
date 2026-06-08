package model

type Paper struct {
	ID              string
	DOI             string
	DisplayName     string
	PublicationDate string
	Authors         []string
	Abstract        string
	LandingPageURL  string
}
