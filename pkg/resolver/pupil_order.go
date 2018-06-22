package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type PupilOrderField int

const (
	PupilPupildAt PupilOrderField = iota
)

func ParsePupilOrderField(s string) (PupilOrderField, error) {
	switch strings.ToUpper(s) {
	case "TUTORED_AT":
		return PupilPupildAt, nil
	default:
		var f PupilOrderField
		return f, fmt.Errorf("invalid PupilOrderField: %q", s)
	}
}

func (f PupilOrderField) String() string {
	switch f {
	case PupilPupildAt:
		return "tutored_at"
	default:
		return "unknown"
	}
}

type PupilOrder struct {
	direction data.OrderDirection
	field     PupilOrderField
}

func NewPupilOrder(d data.OrderDirection, f PupilOrderField) *PupilOrder {
	return &PupilOrder{
		direction: d,
		field:     f,
	}
}

func (o *PupilOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *PupilOrder) Field() string {
	return o.field.String()
}

type PupilOrderArg struct {
	Direction string
	Field     string
}

func ParsePupilOrder(arg *OrderArg) (*PupilOrder, error) {
	if arg == nil {
		return &PupilOrder{
			direction: data.ASC,
			field:     PupilPupildAt,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParsePupilOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	pupilOrder := &PupilOrder{
		direction: direction,
		field:     field,
	}
	return pupilOrder, nil
}

type pupilOrderResolver struct {
	PupilOrder
}

func (r *pupilOrderResolver) Direction() string {
	return r.PupilOrder.Direction().String()
}

func (r *pupilOrderResolver) Field() string {
	return r.PupilOrder.Field()
}
