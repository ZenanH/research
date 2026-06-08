package openalex

import (
	"sort"
	"strings"
)

func ReconstructAbstract(index map[string][]int) string {
	if len(index) == 0 {
		return ""
	}
	positions := make(map[int]string)
	keys := make([]int, 0)
	for word, indexes := range index {
		for _, pos := range indexes {
			positions[pos] = word
			keys = append(keys, pos)
		}
	}
	sort.Ints(keys)
	words := make([]string, 0, len(keys))
	for _, pos := range keys {
		words = append(words, positions[pos])
	}
	return strings.Join(words, " ")
}
