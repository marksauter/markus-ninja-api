package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type LabelableOrderField int

const (
	LabelableCreatedAt LabelableOrderField = iota
	LabelableNumber
	LabelableUpdatedAt
)

func ParseLabelableOrderField(s string) (LabelableOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return LabelableCreatedAt, nil
	case "NUMBER":
		return LabelableNumber, nil
	case "UPDATED_AT":
		return LabelableUpdatedAt, nil
	default:
		var f LabelableOrderField
		return f, fmt.Errorf("invalid LabelableOrderField: %q", s)
	}
}

func (f LabelableOrderField) String() string {
	switch f {
	case LabelableCreatedAt:
		return "created_at"
	case LabelableNumber:
		return "number"
	case LabelableUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type LabelableOrder struct {
	direction data.OrderDirection
	field     LabelableOrderField
}

func (o *LabelableOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *LabelableOrder) Field() string {
	return o.field.String()
}

func ParseLabelableOrder(t LabelableType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &LabelableOrder{
			direction: data.ASC,
			field:     LabelableNumber,
		}, nil
	}
	switch t {
	case LabelableTypeLesson:
		return ParseLessonOrder(arg)
	default:

		return nil, fmt.Errorf("invalid LabelableType: %q", t)
	}
}

type labelableOrderResolver struct {
	data.Order
}

func (r *labelableOrderResolver) Direction() string {
	return r.Order.Direction().String()
}

func (r *labelableOrderResolver) Field() string {
	return r.Order.Field()
}
