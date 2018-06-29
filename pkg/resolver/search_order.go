package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type SearchOrderField int

const (
	SearchAdvancedAt SearchOrderField = iota
	SearchBestMatch
	SearchCreatedAt
	SearchName
	SearchNumber
	SearchUpdatedAt
)

func ParseSearchOrderField(s string) (SearchOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return SearchAdvancedAt, nil
	case "BEST_MATCH":
		return SearchBestMatch, nil
	case "CREATED_AT":
		return SearchCreatedAt, nil
	case "NAME":
		return SearchName, nil
	case "NUMBER":
		return SearchNumber, nil
	case "UPDATED_AT":
		return SearchUpdatedAt, nil
	default:
		var f SearchOrderField
		return f, fmt.Errorf("invalid SearchOrderField: %q", s)
	}
}

func (f SearchOrderField) String() string {
	switch f {
	case SearchAdvancedAt:
		return "advanced_at"
	case SearchBestMatch:
		return "best_match"
	case SearchCreatedAt:
		return "created_at"
	case SearchName:
		return "name"
	case SearchNumber:
		return "number"
	case SearchUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type SearchOrder struct {
	direction data.OrderDirection
	field     SearchOrderField
}

func (o *SearchOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *SearchOrder) Field() string {
	return o.field.String()
}

type OrderArg struct {
	Direction string
	Field     string
}

func ParseSearchOrder(t SearchType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &SearchOrder{
			direction: data.DESC,
			field:     SearchBestMatch,
		}, nil
	}
	switch t {
	case SearchTypeLesson:
		return ParseLessonOrder(arg)
	case SearchTypeStudy:
		return ParseStudyOrder(arg)
	case SearchTypeUser:
		return ParseUserOrder(arg)
	case SearchTypeUserAsset:
		return ParseUserAssetOrder(arg)
	default:

		return nil, fmt.Errorf("invalid SearchType: %q", t)
	}
}

type searchOrderResolver struct {
	data.Order
}

func (r *searchOrderResolver) Direction() string {
	return r.Order.Direction().String()
}

func (r *searchOrderResolver) Field() string {
	return r.Order.Field()
}
