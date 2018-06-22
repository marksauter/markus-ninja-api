package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type FollowingOrderField int

const (
	FollowingFollowingdAt FollowingOrderField = iota
)

func ParseFollowingOrderField(s string) (FollowingOrderField, error) {
	switch strings.ToUpper(s) {
	case "FOLLOWED_AT":
		return FollowingFollowingdAt, nil
	default:
		var f FollowingOrderField
		return f, fmt.Errorf("invalid FollowingOrderField: %q", s)
	}
}

func (f FollowingOrderField) String() string {
	switch f {
	case FollowingFollowingdAt:
		return "followed_at"
	default:
		return "unknown"
	}
}

type FollowingOrder struct {
	direction data.OrderDirection
	field     FollowingOrderField
}

func NewFollowingOrder(d data.OrderDirection, f FollowingOrderField) *FollowingOrder {
	return &FollowingOrder{
		direction: d,
		field:     f,
	}
}

func (o *FollowingOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *FollowingOrder) Field() string {
	return o.field.String()
}

type FollowingOrderArg struct {
	Direction string
	Field     string
}

func ParseFollowingOrder(arg *OrderArg) (*FollowingOrder, error) {
	if arg == nil {
		return &FollowingOrder{
			direction: data.ASC,
			field:     FollowingFollowingdAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseFollowingOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	followingOrder := &FollowingOrder{
		direction: direction,
		field:     field,
	}
	return followingOrder, nil
}

type followingOrderResolver struct {
	FollowingOrder
}

func (r *followingOrderResolver) Direction() string {
	return r.FollowingOrder.Direction().String()
}

func (r *followingOrderResolver) Field() string {
	return r.FollowingOrder.Field()
}
