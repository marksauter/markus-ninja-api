package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type FollowerOrderField int

const (
	FollowerFollowerdAt FollowerOrderField = iota
)

func ParseFollowerOrderField(s string) (FollowerOrderField, error) {
	switch strings.ToUpper(s) {
	case "FOLLOWED_AT":
		return FollowerFollowerdAt, nil
	default:
		var f FollowerOrderField
		return f, fmt.Errorf("invalid FollowerOrderField: %q", s)
	}
}

func (f FollowerOrderField) String() string {
	switch f {
	case FollowerFollowerdAt:
		return "followed_at"
	default:
		return "unknown"
	}
}

type FollowerOrder struct {
	direction data.OrderDirection
	field     FollowerOrderField
}

func NewFollowerOrder(d data.OrderDirection, f FollowerOrderField) *FollowerOrder {
	return &FollowerOrder{
		direction: d,
		field:     f,
	}
}

func (o *FollowerOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *FollowerOrder) Field() string {
	return o.field.String()
}

type FollowerOrderArg struct {
	Direction string
	Field     string
}

func ParseFollowerOrder(arg *OrderArg) (*FollowerOrder, error) {
	if arg == nil {
		return &FollowerOrder{
			direction: data.ASC,
			field:     FollowerFollowerdAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseFollowerOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	followerOrder := &FollowerOrder{
		direction: direction,
		field:     field,
	}
	return followerOrder, nil
}

type followerOrderResolver struct {
	FollowerOrder
}

func (r *followerOrderResolver) Direction() string {
	return r.FollowerOrder.Direction().String()
}

func (r *followerOrderResolver) Field() string {
	return r.FollowerOrder.Field()
}
