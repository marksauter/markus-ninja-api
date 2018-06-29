package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type TopicableOrderField int

const (
	TopicableAdvancedAt TopicableOrderField = iota
	TopicableCreatedAt
	TopicableName
	TopicableUpdatedAt
)

func ParseTopicableOrderField(s string) (TopicableOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return TopicableAdvancedAt, nil
	case "CREATED_AT":
		return TopicableCreatedAt, nil
	case "NAME":
		return TopicableName, nil
	case "UPDATED_AT":
		return TopicableUpdatedAt, nil
	default:
		var f TopicableOrderField
		return f, fmt.Errorf("invalid TopicableOrderField: %q", s)
	}
}

func (f TopicableOrderField) String() string {
	switch f {
	case TopicableAdvancedAt:
		return "advanced_at"
	case TopicableCreatedAt:
		return "created_at"
	case TopicableName:
		return "name"
	case TopicableUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type TopicableOrder struct {
	direction data.OrderDirection
	field     TopicableOrderField
}

func (o *TopicableOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *TopicableOrder) Field() string {
	return o.field.String()
}

func ParseTopicableOrder(t TopicableType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &TopicableOrder{
			direction: data.ASC,
			field:     TopicableName,
		}, nil
	}
	switch t {
	case TopicableTypeStudy:
		return ParseStudyOrder(arg)
	default:
		return nil, fmt.Errorf("invalid TopicableType: %q", t)
	}
}

type topicableOrderResolver struct {
	data.Order
}

func (r *topicableOrderResolver) Direction() string {
	return r.Order.Direction().String()
}

func (r *topicableOrderResolver) Field() string {
	return r.Order.Field()
}
