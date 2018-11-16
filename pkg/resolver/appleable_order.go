package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type AppleableOrderField int

const (
	AppleableAdvancedAt AppleableOrderField = iota
	AppleableAppledAt
	AppleableCreatedAt
	AppleableName
	AppleableUpdatedAt
)

func ParseAppleableOrderField(s string) (AppleableOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return AppleableAdvancedAt, nil
	case "APPLED_AT":
		return AppleableAppledAt, nil
	case "CREATED_AT":
		return AppleableCreatedAt, nil
	case "NAME":
		return AppleableName, nil
	case "UPDATED_AT":
		return AppleableUpdatedAt, nil
	default:
		var f AppleableOrderField
		return f, fmt.Errorf("invalid AppleableOrderField: %q", s)
	}
}

func (f AppleableOrderField) String() string {
	switch f {
	case AppleableAdvancedAt:
		return "advanced_at"
	case AppleableAppledAt:
		return "appled_at"
	case AppleableCreatedAt:
		return "created_at"
	case AppleableName:
		return "name"
	case AppleableUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type AppleableOrder struct {
	direction data.OrderDirection
	field     AppleableOrderField
}

func NewAppleableOrder(d data.OrderDirection, f AppleableOrderField) *AppleableOrder {
	return &AppleableOrder{
		direction: d,
		field:     f,
	}
}

func (o *AppleableOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *AppleableOrder) Field() string {
	return o.field.String()
}

type AppleableOrderArg struct {
	Direction string
	Field     string
}

func ParseAppleableOrder(t AppleableType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &AppleableOrder{
			direction: data.DESC,
			field:     AppleableAppledAt,
		}, nil
	}
	switch t {
	case AppleableTypeCourse:
		return ParseCourseOrder(arg)
	case AppleableTypeStudy:
		return ParseStudyOrder(arg)
	default:
		return nil, fmt.Errorf("invalid AppleableType: %q", t)
	}
}

type appleOrderResolver struct {
	AppleableOrder
}

func (r *appleOrderResolver) Direction() string {
	return r.AppleableOrder.Direction().String()
}

func (r *appleOrderResolver) Field() string {
	return r.AppleableOrder.Field()
}
