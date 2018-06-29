package resolver

import (
	"fmt"
	"strings"
)

type LabelableType int

const (
	LabelableTypeLesson LabelableType = iota
)

func ParseLabelableType(s string) (LabelableType, error) {
	switch strings.ToUpper(s) {
	case "LESSON":
		return LabelableTypeLesson, nil
	default:
		var f LabelableType
		return f, fmt.Errorf("invalid LabelableType: %q", s)
	}
}

func (f LabelableType) String() string {
	switch f {
	case LabelableTypeLesson:
		return "lesson"
	default:
		return "unknown"
	}
}
