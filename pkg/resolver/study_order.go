package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type StudyOrderField int

const (
	StudyAdvancedAt StudyOrderField = iota
	StudyCreatedAt
	StudyName
	StudyUpdatedAt
)

func ParseStudyOrderField(s string) (StudyOrderField, error) {
	switch strings.ToUpper(s) {
	case "ADVANCED_AT":
		return StudyAdvancedAt, nil
	case "CREATED_AT":
		return StudyCreatedAt, nil
	case "NAME":
		return StudyName, nil
	case "UPDATED_AT":
		return StudyUpdatedAt, nil
	default:
		var f StudyOrderField
		return f, fmt.Errorf("invalid StudyOrderField: %q", s)
	}
}

func (f StudyOrderField) String() string {
	switch f {
	case StudyAdvancedAt:
		return "advanced_at"
	case StudyCreatedAt:
		return "created_at"
	case StudyName:
		return "name"
	case StudyUpdatedAt:
		return "updated_at"
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

type StudyOrderArg struct {
	Direction string
	Field     string
}

func ParseStudyOrder(arg *StudyOrderArg) (*StudyOrder, error) {
	if arg == nil {
		arg = &StudyOrderArg{
			Direction: "ASC",
			Field:     "UPDATED_AT",
		}
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
