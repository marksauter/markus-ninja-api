package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type TopicOrderField int

const (
	TopicCreatedAt TopicOrderField = iota
	TopicName
)

func ParseTopicOrderField(s string) (TopicOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return TopicCreatedAt, nil
	case "NAME":
		return TopicName, nil
	default:
		var f TopicOrderField
		return f, fmt.Errorf("invalid TopicOrderField: %q", s)
	}
}

func (f TopicOrderField) String() string {
	switch f {
	case TopicCreatedAt:
		return "created_at"
	case TopicName:
		return "number"
	default:
		return "unknown"
	}
}

type TopicOrder struct {
	direction data.OrderDirection
	field     TopicOrderField
}

func (o *TopicOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *TopicOrder) Field() string {
	return o.field.String()
}

func ParseTopicOrder(arg *OrderArg) (*TopicOrder, error) {
	if arg == nil {
		return &TopicOrder{
			direction: data.DESC,
			field:     TopicCreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseTopicOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	topicOrder := &TopicOrder{
		direction: direction,
		field:     field,
	}
	return topicOrder, nil
}

type topicOrderResolver struct {
	TopicOrder
}

func NewTopicOrder(d data.OrderDirection, f TopicOrderField) *TopicOrder {
	return &TopicOrder{
		direction: d,
		field:     f,
	}
}

func (r *topicOrderResolver) Direction() string {
	return r.TopicOrder.Direction().String()
}

func (r *topicOrderResolver) Field() string {
	return r.TopicOrder.Field()
}