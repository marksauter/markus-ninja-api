package resolver

import (
	"fmt"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type OrderArg struct {
	Direction string
	Field     string
}

func ParseSearchOrder(t SearchType, arg *OrderArg) (data.Order, error) {
	switch t {
	case SearchTypeLesson:
		return ParseLessonOrder(arg)
	case SearchTypeStudy:
		return ParseStudyOrder(arg)
	case SearchTypeUser:
		return ParseUserOrder(arg)
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
