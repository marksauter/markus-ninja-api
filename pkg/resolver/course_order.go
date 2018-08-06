package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type CourseOrderField int

const (
	CourseAdvancedAt CourseOrderField = iota
	CourseAppleCount
	CourseAppledAt
	CourseCreatedAt
	CourseLessonCount
	CourseName
	CourseNumber
)

func ParseCourseOrderField(s string) (CourseOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return CourseAdvancedAt, nil
	case "APPLE_COUNT":
		return CourseAppleCount, nil
	case "APPLED_AT":
		return CourseAppledAt, nil
	case "CREATED_AT":
		return CourseCreatedAt, nil
	case "LESSON_COUNT":
		return CourseLessonCount, nil
	case "NAME":
		return CourseName, nil
	case "NUMBER":
		return CourseNumber, nil
	default:
		var f CourseOrderField
		return f, fmt.Errorf("invalid CourseOrderField: %q", s)
	}
}

func (f CourseOrderField) String() string {
	switch f {
	case CourseAdvancedAt:
		return "advanced_at"
	case CourseAppleCount:
		return "apple_count"
	case CourseAppledAt:
		return "appled_at"
	case CourseCreatedAt:
		return "created_at"
	case CourseLessonCount:
		return "lesson_count"
	case CourseName:
		return "name"
	case CourseNumber:
		return "number"
	default:
		return "unknown"
	}
}

type CourseOrder struct {
	direction data.OrderDirection
	field     CourseOrderField
}

func (o *CourseOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *CourseOrder) Field() string {
	return o.field.String()
}

func ParseCourseOrder(arg *OrderArg) (*CourseOrder, error) {
	if arg == nil {
		return &CourseOrder{
			direction: data.DESC,
			field:     CourseAdvancedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseCourseOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	courseOrder := &CourseOrder{
		direction: direction,
		field:     field,
	}
	return courseOrder, nil
}

type courseOrderResolver struct {
	CourseOrder
}

func (r *courseOrderResolver) Direction() string {
	return r.CourseOrder.Direction().String()
}

func (r *courseOrderResolver) Field() string {
	return r.CourseOrder.Field()
}
