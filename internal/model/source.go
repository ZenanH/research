package model

import "strings"

type Source struct {
	ID          string
	DisplayName string
	ISSNL       string
	ISSN        []string
	WorksCount  int
}

func (s Source) ShortID() string {
	return strings.TrimPrefix(s.ID, "https://openalex.org/")
}
