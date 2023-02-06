package utils

import (
	"sort"
)

func SortStrings(v []string) []string {
	sort.Strings(v)
	return v
}
