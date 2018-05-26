package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type EmailOrderField int

const (
	EmailCreatedAt EmailOrderField = iota
	EmailType
)

func ParseEmailOrderField(s string) (EmailOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return EmailCreatedAt, nil
	case "TYPE":
		return EmailType, nil
	default:
		var f EmailOrderField
		return f, fmt.Errorf("invalid EmailOrderField: %q", s)
	}
}

func (f EmailOrderField) String() string {
	switch f {
	case EmailCreatedAt:
		return "created_at"
	case EmailType:
		return "type"
	default:
		return "unknown"
	}
}

type EmailOrder struct {
	direction data.OrderDirection
	field     EmailOrderField
}

func NewEmailOrder(d data.OrderDirection, f EmailOrderField) *EmailOrder {
	return &EmailOrder{
		direction: d,
		field:     f,
	}
}

func (o *EmailOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *EmailOrder) Field() string {
	return o.field.String()
}

type EmailOrderArg struct {
	Direction string
	Field     string
}

func ParseEmailOrder(arg *EmailOrderArg) (*EmailOrder, error) {
	if arg == nil {
		arg = &EmailOrderArg{
			Direction: "ASC",
			Field:     "CREATED_AT",
		}
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseEmailOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	emailOrder := &EmailOrder{
		direction: direction,
		field:     field,
	}
	return emailOrder, nil
}

type emailOrderResolver struct {
	EmailOrder
}

func (r *emailOrderResolver) Direction() string {
	return r.EmailOrder.Direction().String()
}

func (r *emailOrderResolver) Field() string {
	return r.EmailOrder.Field()
}
