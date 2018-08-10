package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type CommentableOrderField int

const (
	CommentableCreatedAt CommentableOrderField = iota
	CommentableCommentedAt
	CommentableNumber
	CommentableUpdatedAt
)

func ParseCommentableOrderField(s string) (CommentableOrderField, error) {
	switch strings.ToUpper(s) {
	case "COMMENTED_AT":
		return CommentableCommentedAt, nil
	case "CREATED_AT":
		return CommentableCreatedAt, nil
	case "NUMBER":
		return CommentableNumber, nil
	case "UPDATED_AT":
		return CommentableUpdatedAt, nil
	default:
		var f CommentableOrderField
		return f, fmt.Errorf("invalid CommentableOrderField: %q", s)
	}
}

func (f CommentableOrderField) String() string {
	switch f {
	case CommentableCommentedAt:
		return "commented_at"
	case CommentableCreatedAt:
		return "created_at"
	case CommentableNumber:
		return "number"
	case CommentableUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type CommentableOrder struct {
	direction data.OrderDirection
	field     CommentableOrderField
}

func (o *CommentableOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *CommentableOrder) Field() string {
	return o.field.String()
}

func ParseCommentableOrder(t CommentableType, arg *OrderArg) (data.Order, error) {
	if arg == nil {
		return &CommentableOrder{
			direction: data.DESC,
			field:     CommentableCommentedAt,
		}, nil
	}
	switch t {
	case CommentableTypeLesson:
		return ParseLessonOrder(arg)
	default:

		return nil, fmt.Errorf("invalid CommentableType: %q", t)
	}
}

type commentableOrderResolver struct {
	data.Order
}

func (r *commentableOrderResolver) Direction() string {
	return r.Order.Direction().String()
}

func (r *commentableOrderResolver) Field() string {
	return r.Order.Field()
}
