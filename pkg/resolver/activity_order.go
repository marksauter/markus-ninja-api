package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type ActivityOrderField int

const (
	ActivityAdvancedAt ActivityOrderField = iota
	ActivityCreatedAt
	ActivityAssetCount
	ActivityName
	ActivityNumber
)

func ParseActivityOrderField(s string) (ActivityOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return ActivityAdvancedAt, nil
	case "CREATED_AT":
		return ActivityCreatedAt, nil
	case "ASSET_COUNT":
		return ActivityAssetCount, nil
	case "NAME":
		return ActivityName, nil
	case "NUMBER":
		return ActivityNumber, nil
	default:
		var f ActivityOrderField
		return f, fmt.Errorf("invalid ActivityOrderField: %q", s)
	}
}

func (f ActivityOrderField) String() string {
	switch f {
	case ActivityAdvancedAt:
		return "advanced_at"
	case ActivityCreatedAt:
		return "created_at"
	case ActivityAssetCount:
		return "asset_count"
	case ActivityName:
		return "name"
	case ActivityNumber:
		return "number"
	default:
		return "unknown"
	}
}

type ActivityOrder struct {
	direction data.OrderDirection
	field     ActivityOrderField
}

func (o *ActivityOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *ActivityOrder) Field() string {
	return o.field.String()
}

func ParseActivityOrder(arg *OrderArg) (*ActivityOrder, error) {
	if arg == nil {
		return &ActivityOrder{
			direction: data.DESC,
			field:     ActivityAdvancedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseActivityOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	activityOrder := &ActivityOrder{
		direction: direction,
		field:     field,
	}
	return activityOrder, nil
}

type activityOrderResolver struct {
	ActivityOrder
}

func (r *activityOrderResolver) Direction() string {
	return r.ActivityOrder.Direction().String()
}

func (r *activityOrderResolver) Field() string {
	return r.ActivityOrder.Field()
}
