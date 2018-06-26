package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type NotificationOrderField int

const (
	NotificationCreatedAt NotificationOrderField = iota
)

func ParseNotificationOrderField(s string) (NotificationOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return NotificationCreatedAt, nil
	default:
		var f NotificationOrderField
		return f, fmt.Errorf("invalid NotificationOrderField: %q", s)
	}
}

func (f NotificationOrderField) String() string {
	switch f {
	case NotificationCreatedAt:
		return "created_at"
	default:
		return "unknown"
	}
}

type NotificationOrder struct {
	direction data.OrderDirection
	field     NotificationOrderField
}

func NewNotificationOrder(d data.OrderDirection, f NotificationOrderField) *NotificationOrder {
	return &NotificationOrder{
		direction: d,
		field:     f,
	}
}

func (o *NotificationOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *NotificationOrder) Field() string {
	return o.field.String()
}

type NotificationOrderArg struct {
	Direction string
	Field     string
}

func ParseNotificationOrder(arg *OrderArg) (*NotificationOrder, error) {
	if arg == nil {
		return &NotificationOrder{
			direction: data.DESC,
			field:     NotificationCreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseNotificationOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	notificationOrder := &NotificationOrder{
		direction: direction,
		field:     field,
	}
	return notificationOrder, nil
}

type notificationOrderResolver struct {
	NotificationOrder
}

func (r *notificationOrderResolver) Direction() string {
	return r.NotificationOrder.Direction().String()
}

func (r *notificationOrderResolver) Field() string {
	return r.NotificationOrder.Field()
}
