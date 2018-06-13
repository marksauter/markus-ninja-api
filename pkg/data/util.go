package data

import (
	"strings"

	"github.com/jackc/pgx/pgtype"
)

func ToTsQuery(query string) string {
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
