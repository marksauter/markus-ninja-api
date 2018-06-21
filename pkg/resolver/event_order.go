package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type EventOrderField int

const (
	CreatedAt EventOrderField = iota
)

func ParseEventOrderField(s string) (EventOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return CreatedAt, nil
	default:
		var f EventOrderField
		return f, fmt.Errorf("invalid EventOrderField: %q", s)
	}
}

func (f EventOrderField) String() string {
	switch f {
	case CreatedAt:
		return "created_at"
	default:
		return "unknown"
	}
}

type EventOrder struct {
	direction data.OrderDirection
	field     EventOrderField
}

func NewEventOrder(d data.OrderDirection, f EventOrderField) *EventOrder {
	return &EventOrder{
		direction: d,
		field:     f,
	}
}

func (o *EventOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *EventOrder) Field() string {
	return o.field.String()
}

type EventOrderArg struct {
	Direction string
	Field     string
}

func ParseEventOrder(arg *OrderArg) (*EventOrder, error) {
	if arg == nil {
		return &EventOrder{
			direction: data.DESC,
			field:     CreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseEventOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	eventOrder := &EventOrder{
		direction: direction,
		field:     field,
	}
	return eventOrder, nil
}

type eventOrderResolver struct {
	EventOrder
}

func (r *eventOrderResolver) Direction() string {
	return r.EventOrder.Direction().String()
}

func (r *eventOrderResolver) Field() string {
	return r.EventOrder.Field()
}
