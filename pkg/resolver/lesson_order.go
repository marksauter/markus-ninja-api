package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type LessonOrderField int

const (
	LessonCreatedAt LessonOrderField = iota
	LessonCommentCount
	LessonCourseNumber
	LessonNumber
	LessonPublishedAt
	LessonUpdatedAt
)

func ParseLessonOrderField(s string) (LessonOrderField, error) {
	switch strings.ToUpper(s) {
	case "CREATED_AT":
		return LessonCreatedAt, nil
	case "COMMENT_COUNT":
		return LessonCommentCount, nil
	case "COURSE_NUMBER":
		return LessonCourseNumber, nil
	case "NUMBER":
		return LessonNumber, nil
	case "PUBLISHED_AT":
		return LessonPublishedAt, nil
	case "UPDATED_AT":
		return LessonUpdatedAt, nil
	default:
		var f LessonOrderField
		return f, fmt.Errorf("invalid LessonOrderField: %q", s)
	}
}

func (f LessonOrderField) String() string {
	switch f {
	case LessonCreatedAt:
		return "created_at"
	case LessonCommentCount:
		return "comment_count"
	case LessonCourseNumber:
		return "course_number"
	case LessonNumber:
		return "number"
	case LessonPublishedAt:
		return "published_at"
	case LessonUpdatedAt:
		return "updated_at"
	default:
		return "unknown"
	}
}

type LessonOrder struct {
	direction data.OrderDirection
	field     LessonOrderField
}

func (o *LessonOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *LessonOrder) Field() string {
	return o.field.String()
}

func ParseLessonOrder(arg *OrderArg) (*LessonOrder, error) {
	if arg == nil {
		return &LessonOrder{
			direction: data.ASC,
			field:     LessonNumber,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseLessonOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	lessonOrder := &LessonOrder{
		direction: direction,
		field:     field,
	}
	return lessonOrder, nil
}

type lessonOrderResolver struct {
	LessonOrder
}

func NewLessonOrder(d data.OrderDirection, f LessonOrderField) *LessonOrder {
	return &LessonOrder{
		direction: d,
		field:     f,
	}
}

func (r *lessonOrderResolver) Direction() string {
	return r.LessonOrder.Direction().String()
}

func (r *lessonOrderResolver) Field() string {
	return r.LessonOrder.Field()
}
