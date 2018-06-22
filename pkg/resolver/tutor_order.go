package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type TutorOrderField int

const (
	TutorTutordAt TutorOrderField = iota
)

func ParseTutorOrderField(s string) (TutorOrderField, error) {
	switch strings.ToUpper(s) {
	case "TUTORED_AT":
		return TutorTutordAt, nil
	default:
		var f TutorOrderField
		return f, fmt.Errorf("invalid TutorOrderField: %q", s)
	}
}

func (f TutorOrderField) String() string {
	switch f {
	case TutorTutordAt:
		return "enrolled_at"
	default:
		return "unknown"
	}
}

type TutorOrder struct {
	direction data.OrderDirection
	field     TutorOrderField
}

func NewTutorOrder(d data.OrderDirection, f TutorOrderField) *TutorOrder {
	return &TutorOrder{
		direction: d,
		field:     f,
	}
}

func (o *TutorOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *TutorOrder) Field() string {
	return o.field.String()
}

type TutorOrderArg struct {
	Direction string
	Field     string
}

func ParseTutorOrder(arg *OrderArg) (*TutorOrder, error) {
	if arg == nil {
		return &TutorOrder{
			direction: data.ASC,
			field:     TutorTutordAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseTutorOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	tutorOrder := &TutorOrder{
		direction: direction,
		field:     field,
	}
	return tutorOrder, nil
}

type tutorOrderResolver struct {
	TutorOrder
}

func (r *tutorOrderResolver) Direction() string {
	return r.TutorOrder.Direction().String()
}

func (r *tutorOrderResolver) Field() string {
	return r.TutorOrder.Field()
}
