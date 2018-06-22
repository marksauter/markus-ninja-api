package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type StudentOrderField int

const (
	StudentEnrolledAt StudentOrderField = iota
)

func ParseStudentOrderField(s string) (StudentOrderField, error) {
	switch strings.ToUpper(s) {
	case "ENROLLED_AT":
		return StudentEnrolledAt, nil
	default:
		var f StudentOrderField
		return f, fmt.Errorf("invalid StudentOrderField: %q", s)
	}
}

func (f StudentOrderField) String() string {
	switch f {
	case StudentEnrolledAt:
		return "enrolled_at"
	default:
		return "unknown"
	}
}

type StudentOrder struct {
	direction data.OrderDirection
	field     StudentOrderField
}

func NewStudentOrder(d data.OrderDirection, f StudentOrderField) *StudentOrder {
	return &StudentOrder{
		direction: d,
		field:     f,
	}
}

func (o *StudentOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *StudentOrder) Field() string {
	return o.field.String()
}

type StudentOrderArg struct {
	Direction string
	Field     string
}

func ParseStudentOrder(arg *OrderArg) (*StudentOrder, error) {
	if arg == nil {
		return &StudentOrder{
			direction: data.ASC,
			field:     StudentEnrolledAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseStudentOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	studentOrder := &StudentOrder{
		direction: direction,
		field:     field,
	}
	return studentOrder, nil
}

type studentOrderResolver struct {
	StudentOrder
}

func (r *studentOrderResolver) Direction() string {
	return r.StudentOrder.Direction().String()
}

func (r *studentOrderResolver) Field() string {
	return r.StudentOrder.Field()
}
