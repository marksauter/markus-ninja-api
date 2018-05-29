package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type LessonCommentOrderField int

const (
	LessonCommentCreatedAt LessonCommentOrderField = iota
	LessonCommentUpdatedAt
)

func ParseLessonCommentOrderField(s string) (LessonCommentOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return LessonCommentCreatedAt, nil
	case "UPDATED_AT":
		return LessonCommentUpdatedAt, nil
	default:
		var f LessonCommentOrderField
		return f, fmt.Errorf("invalid LessonCommentOrderField: %q", s)
	}
}

func (f LessonCommentOrderField) String() string {
	switch f {
	case LessonCommentCreatedAt:
		return "created_at"
	case LessonCommentUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type LessonCommentOrder struct {
	direction data.OrderDirection
	field     LessonCommentOrderField
}

func (o *LessonCommentOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *LessonCommentOrder) Field() string {
	return o.field.String()
}

type LessonCommentOrderArg struct {
	Direction string
	Field     string
}

func ParseLessonCommentOrder(arg *LessonCommentOrderArg) (*LessonCommentOrder, error) {
	if arg == nil {
		arg = &LessonCommentOrderArg{
			Direction: "ASC",
			Field:     "CREATED_AT",
		}
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseLessonCommentOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	lessonCommentOrder := &LessonCommentOrder{
		direction: direction,
		field:     field,
	}
	return lessonCommentOrder, nil
}

type lessonCommentOrderResolver struct {
	LessonCommentOrder
}

func NewLessonCommentOrder(
	d data.OrderDirection,
	f LessonCommentOrderField,
) *LessonCommentOrder {
	return &LessonCommentOrder{
		direction: d,
		field:     f,
	}
}

func (r *lessonCommentOrderResolver) Direction() string {
	return r.LessonCommentOrder.Direction().String()
}

func (r *lessonCommentOrderResolver) Field() string {
	return r.LessonCommentOrder.Field()
}
