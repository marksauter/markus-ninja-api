package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type EnrolleeOrderField int

const (
	EnrolleeEnrolledAt EnrolleeOrderField = iota
)

func ParseEnrolleeOrderField(s string) (EnrolleeOrderField, error) {
	switch strings.ToUpper(s) {
	case "ENROLLED_AT":
		return EnrolleeEnrolledAt, nil
	default:
		var f EnrolleeOrderField
		return f, fmt.Errorf("invalid EnrolleeOrderField: %q", s)
	}
}

func (f EnrolleeOrderField) String() string {
	switch f {
	case EnrolleeEnrolledAt:
		return "enrolled_at"
	default:
		return "unknown"
	}
}

type EnrolleeOrder struct {
	direction data.OrderDirection
	field     EnrolleeOrderField
}

func NewEnrolleeOrder(d data.OrderDirection, f EnrolleeOrderField) *EnrolleeOrder {
	return &EnrolleeOrder{
		direction: d,
		field:     f,
	}
}

func (o *EnrolleeOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *EnrolleeOrder) Field() string {
	return o.field.String()
}

type EnrolleeOrderArg struct {
	Direction string
	Field     string
}

func ParseEnrolleeOrder(arg *OrderArg) (*EnrolleeOrder, error) {
	if arg == nil {
		return &EnrolleeOrder{
			direction: data.ASC,
			field:     EnrolleeEnrolledAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseEnrolleeOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	enrolleeOrder := &EnrolleeOrder{
		direction: direction,
		field:     field,
	}
	return enrolleeOrder, nil
}

type enrolleeOrderResolver struct {
	EnrolleeOrder
}

func (r *enrolleeOrderResolver) Direction() string {
	return r.EnrolleeOrder.Direction().String()
}

func (r *enrolleeOrderResolver) Field() string {
	return r.EnrolleeOrder.Field()
}
