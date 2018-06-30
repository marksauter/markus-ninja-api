package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type EnrollableOrderField int

const (
	EnrollableAdvancedAt EnrollableOrderField = iota
	EnrollableCreatedAt
	EnrollableEnrolledAt
	EnrollableName
	EnrollableUpdatedAt
)

func ParseEnrollableOrderField(s string) (EnrollableOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return EnrollableAdvancedAt, nil
	case "CREATED_AT":
		return EnrollableCreatedAt, nil
	case "ENROLLED_AT":
		return EnrollableEnrolledAt, nil
	case "NAME":
		return EnrollableName, nil
	case "UPDATED_AT":
		return EnrollableUpdatedAt, nil
	default:
		var f EnrollableOrderField
		return f, fmt.Errorf("invalid EnrollableOrderField: %q", s)
	}
}

func (f EnrollableOrderField) String() string {
	switch f {
	case EnrollableAdvancedAt:
		return "advanced_at"
	case EnrollableCreatedAt:
		return "created_at"
	case EnrollableEnrolledAt:
		return "enrolled_at"
	case EnrollableName:
		return "name"
	case EnrollableUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type EnrollableOrder struct {
	direction data.OrderDirection
	field     EnrollableOrderField
}

func (o *EnrollableOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *EnrollableOrder) Field() string {
	return o.field.String()
}

func ParseEnrollableOrder(t EnrollableType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &EnrollableOrder{
			direction: data.DESC,
			field:     EnrollableEnrolledAt,
		}, nil
	}
	switch t {
	case EnrollableTypeStudy:
		return ParseStudyOrder(arg)
	default:
		return nil, fmt.Errorf("invalid EnrollableType: %q", t)
	}
}

type enrollableOrderResolver struct {
	data.Order
}

func (r *enrollableOrderResolver) Direction() string {
	return r.Order.Direction().String()
}

func (r *enrollableOrderResolver) Field() string {
	return r.Order.Field()
}
