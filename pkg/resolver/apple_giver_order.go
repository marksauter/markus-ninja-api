package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type AppleGiverOrderField int

const (
	AppleGiverAdvancedAt AppleGiverOrderField = iota
	AppleGiverAppledAt
)

func ParseAppleGiverOrderField(s string) (AppleGiverOrderField, error) {
	switch strings.ToUpper(s) {
	case "APPLED_AT":
		return AppleGiverAppledAt, nil
	default:
		var f AppleGiverOrderField
		return f, fmt.Errorf("invalid AppleGiverOrderField: %q", s)
	}
}

func (f AppleGiverOrderField) String() string {
	switch f {
	case AppleGiverAppledAt:
		return "appled_at"
	default:
		return "unknown"
	}
}

type AppleGiverOrder struct {
	direction data.OrderDirection
	field     AppleGiverOrderField
}

func NewAppleGiverOrder(d data.OrderDirection, f AppleGiverOrderField) *AppleGiverOrder {
	return &AppleGiverOrder{
		direction: d,
		field:     f,
	}
}

func (o *AppleGiverOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *AppleGiverOrder) Field() string {
	return o.field.String()
}

type AppleGiverOrderArg struct {
	Direction string
	Field     string
}

func ParseAppleGiverOrder(arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &AppleGiverOrder{
			direction: data.DESC,
			field:     AppleGiverAppledAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseAppleGiverOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	appleGiverOrder := &AppleGiverOrder{
		direction: direction,
		field:     field,
	}
	return appleGiverOrder, nil
}

type appleGiverOrderResolver struct {
	AppleGiverOrder
}

func (r *appleGiverOrderResolver) Direction() string {
	return r.AppleGiverOrder.Direction().String()
}

func (r *appleGiverOrderResolver) Field() string {
	return r.AppleGiverOrder.Field()
}
