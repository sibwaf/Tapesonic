package util

import (
	"fmt"
	"strings"
)

func MakeTextSearchCondition(fields []string, query string) string {
	terms := extractSearchTerms(query)
	if len(terms) == 0 {
		return ""
	}

	searchField := "''"
	for _, field := range fields {
		searchField = fmt.Sprintf("%s || ' ' || coalesce(%s, '')", searchField, field)
	}

	filter := []string{}
	for _, term := range terms {
		term = escapeTextLiteralForLike(term, "\\")
		filter = append(filter, fmt.Sprintf("%s LIKE '%% %s%%' ESCAPE '%s'", searchField, term, "\\"))
	}

	return strings.Join(filter, " AND ")
}

func MakeTextSearchString(raw string) string {
	return strings.Join(extractSearchTerms(raw), " ")
}

func extractSearchTerms(query string) []string {
	query, err := NormalizeUnicode(query)
	if err != nil {
		return []string{}
	}

	return SplitWords(query)
}

func escapeTextLiteralForLike(str string, escape string) string {
	str = escapeTextLiteral(str)
	str = strings.ReplaceAll(str, escape, escape+escape)
	str = strings.ReplaceAll(str, "_", escape+"_")
	str = strings.ReplaceAll(str, "%", escape+"%")
	return str
}

func escapeTextLiteral(str string) string {
	return strings.ReplaceAll(str, "'", "''")
}
