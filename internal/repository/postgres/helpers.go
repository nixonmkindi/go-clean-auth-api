package postgres

import "strings"

func sanitizeSort(sortBy, sortOrder string, allowed map[string]string, defaultColumn string) (string, string) {
	column, ok := allowed[strings.ToLower(sortBy)]
	if !ok {
		column = defaultColumn
	}

	order := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		order = "ASC"
	}

	return column, order
}
