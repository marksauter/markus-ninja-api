package data

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

func ToSubQuery(query string) string {
	return fmt.Sprintf("(%s)", query)
}

func ToTsQuery(query string) string {
	if strings.TrimSpace(query) == "" {
		return "*"
	}
	words := strings.FieldsFunc(query, queryDelimiter)
	return strings.Join(words, " & ")
}

func ToPrefixTsQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" || query == "*" {
		return "*"
	}
	words := strings.FieldsFunc(query, queryDelimiter)
	for i, v := range words {
		words[i] = v + ":*"
	}
	return strings.Join(words, " & ")
}

func ToLikeAnyPatternQuery(query string) *pgtype.TextArray {
	words := strings.FieldsFunc(query, queryDelimiter)
	for i, v := range words {
		words[i] = strings.ToLower(v + "%")
	}
	dbArray := &pgtype.TextArray{}
	dbArray.Set(words)
	return dbArray
}

func queryDelimiter(r rune) bool {
	return r == ' ' ||
		r == '_' ||
		r == '-'
}
