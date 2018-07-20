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
	case "BEST_MATCH":
		return SearchBestMatch, nil
	default:
		var f SearchOrderField
		return f, fmt.Errorf("invalid SearchOrderField: %q", s)
	}
}

func (f SearchOrderField) String() string {
	switch f {
	case SearchBestMatch:
		return "best_match"
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
	if field, err := ParseSearchOrderField(arg.Field); err == nil {
		direction, err := data.ParseOrderDirection(arg.Direction)
		if err != nil {
			return nil, err
		}
		return &SearchOrder{
			direction,
			field,
		}, nil
	}
	switch t {
	case SearchTypeLesson:
		return ParseLessonOrder(arg)
	case SearchTypeStudy:
		return ParseStudyOrder(arg)
	case SearchTypeTopic:
		return ParseTopicOrder(arg)
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
