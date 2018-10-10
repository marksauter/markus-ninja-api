package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type TopicableOrderField int

const (
	TopicableTopicedAt TopicableOrderField = iota
)

func ParseTopicableOrderField(s string) (TopicableOrderField, error) {
	switch strings.ToUpper(s) {
	case "TOPICED_AT":
		return TopicableTopicedAt, nil
	default:
		var f TopicableOrderField
		return f, fmt.Errorf("invalid TopicableOrderField: %q", s)
	}
}

func (f TopicableOrderField) String() string {
	switch f {
	case TopicableTopicedAt:
		return "topiced_at"
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
			direction: data.DESC,
			field:     TopicableTopicedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseTopicableOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	topicableOrder := &TopicableOrder{
		direction: direction,
		field:     field,
	}
	return topicableOrder, nil
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
