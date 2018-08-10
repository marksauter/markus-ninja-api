package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type UserAssetCommentOrderField int

const (
	UserAssetCommentCreatedAt UserAssetCommentOrderField = iota
	UserAssetCommentUpdatedAt
)

func ParseUserAssetCommentOrderField(s string) (UserAssetCommentOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return UserAssetCommentCreatedAt, nil
	case "UPDATED_AT":
		return UserAssetCommentUpdatedAt, nil
	default:
		var f UserAssetCommentOrderField
		return f, fmt.Errorf("invalid UserAssetCommentOrderField: %q", s)
	}
}

func (f UserAssetCommentOrderField) String() string {
	switch f {
	case UserAssetCommentCreatedAt:
		return "created_at"
	case UserAssetCommentUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type UserAssetCommentOrder struct {
	direction data.OrderDirection
	field     UserAssetCommentOrderField
}

func (o *UserAssetCommentOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *UserAssetCommentOrder) Field() string {
	return o.field.String()
}

type UserAssetCommentOrderArg struct {
	Direction string
	Field     string
}

func ParseUserAssetCommentOrder(arg *OrderArg) (*UserAssetCommentOrder, error) {
	if arg == nil {
		return &UserAssetCommentOrder{
			direction: data.ASC,
			field:     UserAssetCommentCreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseUserAssetCommentOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	userAssetCommentOrder := &UserAssetCommentOrder{
		direction: direction,
		field:     field,
	}
	return userAssetCommentOrder, nil
}

type userAssetCommentOrderResolver struct {
	UserAssetCommentOrder
}

func NewUserAssetCommentOrder(
	d data.OrderDirection,
	f UserAssetCommentOrderField,
) *UserAssetCommentOrder {
	return &UserAssetCommentOrder{
		direction: d,
		field:     f,
	}
}

func (r *userAssetCommentOrderResolver) Direction() string {
	return r.UserAssetCommentOrder.Direction().String()
}

func (r *userAssetCommentOrderResolver) Field() string {
	return r.UserAssetCommentOrder.Field()
}
