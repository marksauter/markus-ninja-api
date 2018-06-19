package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type RefOrderField int

const (
	ReferredAt RefOrderField = iota
)

func ParseRefOrderField(s string) (RefOrderField, error) {
	switch strings.ToUpper(s) {
	case "REFERRED_AT":
		return ReferredAt, nil
	default:
		var f RefOrderField
		return f, fmt.Errorf("invalid RefOrderField: %q", s)
	}
}

func (f RefOrderField) String() string {
	switch f {
	case ReferredAt:
		return "referred_at"
	default:
		return "unknown"
	}
}

type RefOrder struct {
	direction data.OrderDirection
	field     RefOrderField
}

func NewRefOrder(d data.OrderDirection, f RefOrderField) *RefOrder {
	return &RefOrder{
		direction: d,
		field:     f,
	}
}

func (o *RefOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *RefOrder) Field() string {
	return o.field.String()
}

type RefOrderArg struct {
	Direction string
	Field     string
}

func ParseRefOrder(arg *OrderArg) (*RefOrder, error) {
	if arg == nil {
		return &RefOrder{
			direction: data.DESC,
			field:     ReferredAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseRefOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	refOrder := &RefOrder{
		direction: direction,
		field:     field,
	}
	return refOrder, nil
}

type refOrderResolver struct {
	RefOrder
}

func (r *refOrderResolver) Direction() string {
	return r.RefOrder.Direction().String()
}

func (r *refOrderResolver) Field() string {
	return r.RefOrder.Field()
}
