package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type UserAssetOrderField int

const (
	UserAssetCreatedAt UserAssetOrderField = iota
	UserAssetName
	UserAssetUpdatedAt
)

func ParseUserAssetOrderField(s string) (UserAssetOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return UserAssetCreatedAt, nil
	case "NAME":
		return UserAssetName, nil
	case "UPDATED_AT":
		return UserAssetUpdatedAt, nil
	default:
		var f UserAssetOrderField
		return f, fmt.Errorf("invalid UserAssetOrderField: %q", s)
	}
}

func (f UserAssetOrderField) String() string {
	switch f {
	case UserAssetCreatedAt:
		return "created_at"
	case UserAssetName:
		return "name"
	case UserAssetUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type UserAssetOrder struct {
	direction data.OrderDirection
	field     UserAssetOrderField
}

func (o *UserAssetOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *UserAssetOrder) Field() string {
	return o.field.String()
}

func ParseUserAssetOrder(arg *OrderArg) (*UserAssetOrder, error) {
	if arg == nil {
		return &UserAssetOrder{
			direction: data.DESC,
			field:     UserAssetCreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseUserAssetOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	userAssetOrder := &UserAssetOrder{
		direction: direction,
		field:     field,
	}
	return userAssetOrder, nil
}

type userAssetOrderResolver struct {
	UserAssetOrder
}

func NewUserAssetOrder(
	d data.OrderDirection,
	f UserAssetOrderField,
) *UserAssetOrder {
	return &UserAssetOrder{
		direction: d,
		field:     f,
	}
}

func (r *userAssetOrderResolver) Direction() string {
	return r.UserAssetOrder.Direction().String()
}

func (r *userAssetOrderResolver) Field() string {
	return r.UserAssetOrder.Field()
}
