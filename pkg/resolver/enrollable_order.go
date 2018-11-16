package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type EnrollableOrderField int

const (
	EnrollableEnrolledAt EnrollableOrderField = iota
)

func ParseEnrollableOrderField(s string) (EnrollableOrderField, error) {
	switch strings.ToUpper(s) {
	case "ENROLLED_AT":
		return EnrollableEnrolledAt, nil
	default:
		var f EnrollableOrderField
		return f, fmt.Errorf("invalid EnrollableOrderField: %q", s)
	}
}

func (f EnrollableOrderField) String() string {
	switch f {
	case EnrollableEnrolledAt:
		return "enrolled_at"
	default:
		return "unknown"
	}
}

type EnrollableOrder struct {
	direction data.OrderDirection
	field     EnrollableOrderField
}

func (o *EnrollableOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *EnrollableOrder) Field() string {
	return o.field.String()
}

func ParseEnrollableOrder(t EnrollableType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &EnrollableOrder{
			direction: data.DESC,
			field:     EnrollableEnrolledAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseEnrollableOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	enrollableOrder := &EnrollableOrder{
		direction: direction,
		field:     field,
	}
	return enrollableOrder, nil
}

type enrollableOrderResolver struct {
	data.Order
}

func (r *enrollableOrderResolver) Direction() string {
	return r.Order.Direction().String()
}

func (r *enrollableOrderResolver) Field() string {
	return r.Order.Field()
}
