package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type AppleOrderField int

const (
	AppleAppledAt AppleOrderField = iota
)

func ParseAppleOrderField(s string) (AppleOrderField, error) {
	switch strings.ToUpper(s) {
	case "APPLED_AT":
		return AppleAppledAt, nil
	default:
		var f AppleOrderField
		return f, fmt.Errorf("invalid AppleOrderField: %q", s)
	}
}

func (f AppleOrderField) String() string {
	switch f {
	case AppleAppledAt:
		return "appled_at"
	default:
		return "unknown"
	}
}

type AppleOrder struct {
	direction data.OrderDirection
	field     AppleOrderField
}

func NewAppleOrder(d data.OrderDirection, f AppleOrderField) *AppleOrder {
	return &AppleOrder{
		direction: d,
		field:     f,
	}
}

func (o *AppleOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *AppleOrder) Field() string {
	return o.field.String()
}

type AppleOrderArg struct {
	Direction string
	Field     string
}

func ParseAppleOrder(arg *OrderArg) (*AppleOrder, error) {
	if arg == nil {
		return &AppleOrder{
			direction: data.ASC,
			field:     AppleAppledAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseAppleOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	appleOrder := &AppleOrder{
		direction: direction,
		field:     field,
	}
	return appleOrder, nil
}

type appleOrderResolver struct {
	AppleOrder
}

func (r *appleOrderResolver) Direction() string {
	return r.AppleOrder.Direction().String()
}

func (r *appleOrderResolver) Field() string {
	return r.AppleOrder.Field()
}
