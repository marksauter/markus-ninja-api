package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type CommentOrderField int

const (
	CommentCreatedAt CommentOrderField = iota
	CommentUpdatedAt
)

func ParseCommentOrderField(s string) (CommentOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return CommentCreatedAt, nil
	case "UPDATED_AT":
		return CommentUpdatedAt, nil
	default:
		var f CommentOrderField
		return f, fmt.Errorf("invalid CommentOrderField: %q", s)
	}
}

func (f CommentOrderField) String() string {
	switch f {
	case CommentCreatedAt:
		return "created_at"
	case CommentUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type CommentOrder struct {
	direction data.OrderDirection
	field     CommentOrderField
}

func (o *CommentOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *CommentOrder) Field() string {
	return o.field.String()
}

type CommentOrderArg struct {
	Direction string
	Field     string
}

func ParseCommentOrder(arg *OrderArg) (*CommentOrder, error) {
	if arg == nil {
		return &CommentOrder{
			direction: data.ASC,
			field:     CommentCreatedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseCommentOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	commentOrder := &CommentOrder{
		direction: direction,
		field:     field,
	}
	return commentOrder, nil
}

type commentOrderResolver struct {
	CommentOrder
}

func NewCommentOrder(
	d data.OrderDirection,
	f CommentOrderField,
) *CommentOrder {
	return &CommentOrder{
		direction: d,
		field:     f,
	}
}

func (r *commentOrderResolver) Direction() string {
	return r.CommentOrder.Direction().String()
}

func (r *commentOrderResolver) Field() string {
	return r.CommentOrder.Field()
}
