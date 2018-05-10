package resolver

import "github.com/marksauter/markus-ninja-api/pkg/data"

type LessonOrder struct {
	Direction data.OrderDirection
	Field     data.LessonOrderField
}

type LessonOrderArg struct {
	Direction string
	Field     string
}

func ParseLessonOrder(arg *LessonOrderArg) (*LessonOrder, error) {
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := data.ParseLessonOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	lessonOrder := &LessonOrder{
		Direction: direction,
		Field:     field,
	}
	return lessonOrder, nil
}

type lessonOrderResolver struct {
	LessonOrder
}

func (r *lessonOrderResolver) Direction() string {
	return r.LessonOrder.Direction.String()
}

func (r *lessonOrderResolver) Field() string {
	return r.LessonOrder.Field.Name()
}
