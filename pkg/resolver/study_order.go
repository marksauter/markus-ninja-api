package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type StudyOrderField int

const (
	StudyAdvancedAt StudyOrderField = iota
	StudyAppleCount
	StudyAppledAt
	StudyCreatedAt
	StudyEnrolledAt
	StudyLessonCount
)

func ParseStudyOrderField(s string) (StudyOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return StudyAdvancedAt, nil
	case "APPLE_COUNT":
		return StudyAppleCount, nil
	case "APPLED_AT":
		return StudyAppledAt, nil
	case "CREATED_AT":
		return StudyCreatedAt, nil
	case "ENROLLED_AT":
		return StudyEnrolledAt, nil
	case "LESSON_COUNT":
		return StudyLessonCount, nil
	default:
		var f StudyOrderField
		return f, fmt.Errorf("invalid StudyOrderField: %q", s)
	}
}

func (f StudyOrderField) String() string {
	switch f {
	case StudyAdvancedAt:
		return "advanced_at"
	case StudyAppleCount:
		return "apple_count"
	case StudyAppledAt:
		return "appled_at"
	case StudyCreatedAt:
		return "created_at"
	case StudyEnrolledAt:
		return "enrolled_at"
	case StudyLessonCount:
		return "lesson_count"
	default:
		return "unknown"
	}
}

type StudyOrder struct {
	direction data.OrderDirection
	field     StudyOrderField
}

func (o *StudyOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *StudyOrder) Field() string {
	return o.field.String()
}

func ParseStudyOrder(arg *OrderArg) (*StudyOrder, error) {
	if arg == nil {
		return &StudyOrder{
			direction: data.DESC,
			field:     StudyAdvancedAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseStudyOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	studyOrder := &StudyOrder{
		direction: direction,
		field:     field,
	}
	return studyOrder, nil
}

type studyOrderResolver struct {
	StudyOrder
}

func (r *studyOrderResolver) Direction() string {
	return r.StudyOrder.Direction().String()
}

func (r *studyOrderResolver) Field() string {
	return r.StudyOrder.Field()
}
