package resolver

import (
	"fmt"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

type LabelOrderField int

const (
	LabelLabeledAt LabelOrderField = iota
	LabelName
)

func ParseLabelOrderField(s string) (LabelOrderField, error) {
	switch strings.ToUpper(s) {
	case "LABELED_AT":
		return LabelLabeledAt, nil
	case "NAME":
		return LabelName, nil
	default:
		var f LabelOrderField
		return f, fmt.Errorf("invalid LabelOrderField: %q", s)
	}
}

func (f LabelOrderField) String() string {
	switch f {
	case LabelLabeledAt:
		return "labeled_at"
	case LabelName:
		return "name"
	default:
		return "unknown"
	}
}

type LabelOrder struct {
	direction data.OrderDirection
	field     LabelOrderField
}

func (o *LabelOrder) Direction() data.OrderDirection {
	return o.direction
}

func (o *LabelOrder) Field() string {
	return o.field.String()
}

func ParseLabelOrder(arg *OrderArg) (*LabelOrder, error) {
	if arg == nil {
		return &LabelOrder{
			direction: data.ASC,
			field:     LabelName,
		}, nil
	}
	direction, err := data.ParseOrderDirection(arg.Direction)
	if err != nil {
		return nil, err
	}
	field, err := ParseLabelOrderField(arg.Field)
	if err != nil {
		return nil, err
	}
	labelOrder := &LabelOrder{
		direction: direction,
		field:     field,
	}
	return labelOrder, nil
}

type labelOrderResolver struct {
	LabelOrder
}

func NewLabelOrder(d data.OrderDirection, f LabelOrderField) *LabelOrder {
	return &LabelOrder{
		direction: d,
		field:     f,
	}
}

func (r *labelOrderResolver) Direction() string {
	return r.LabelOrder.Direction().String()
}

func (r *labelOrderResolver) Field() string {
	return r.LabelOrder.Field()
}
