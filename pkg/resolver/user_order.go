package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type UserOrderField int

const (
	UserCreatedAt UserOrderField = iota
	UserEnrolledAt
	UserEnrolleeCount
	UserStudyCount
)

func ParseUserOrderField(s string) (UserOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return UserCreatedAt, nil
	case "ENROLLED_AT":
		return UserEnrolledAt, nil
	case "ENROLLEE_COUNT":
		return UserEnrolleeCount, nil
	case "STUDY_COUNT":
		return UserStudyCount, nil
	default:
		var f UserOrderField
		return f, fmt.Errorf("invalid UserOrderField: %q", s)
	}
}

func (f UserOrderField) String() string {
	switch f {
	case UserCreatedAt:
		return "created_at"
	case UserEnrolledAt:
		return "enrolled_at"
	case UserEnrolleeCount:
		return "enrollee_count"
	case UserStudyCount:
		return "study_count"
	default:
		return "unknown"
	}
}

type UserOrder struct {
	direction data.OrderDirection
	field     UserOrderField
}

func (o *UserOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *UserOrder) Field() string {
	return o.field.String()
}

func ParseUserOrder(arg *OrderArg) (*UserOrder, error) {
	if arg == nil {
		return &UserOrder{
			direction: data.DESC,
			field:     UserCreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseUserOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	userOrder := &UserOrder{
		direction: direction,
		field:     field,
	}
	return userOrder, nil
}

type userOrderResolver struct {
	UserOrder
}

func NewUserOrder(d data.OrderDirection, f UserOrderField) *UserOrder {
	return &UserOrder{
		direction: d,
		field:     f,
	}
}

func (r *userOrderResolver) Direction() string {
	return r.UserOrder.Direction().String()
}

func (r *userOrderResolver) Field() string {
	return r.UserOrder.Field()
}
