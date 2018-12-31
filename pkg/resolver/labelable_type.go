package resolver

import (
	"fmt"
	"strings"
)

type LabelableType int

const (
	LabelableTypeComment LabelableType = iota
	LabelableTypeLesson
	LabelableTypeUserAsset
)

func ParseLabelableType(s string) (LabelableType, error) {
	switch strings.ToUpper(s) {
	case "COMMENT":
		return LabelableTypeComment, nil
	case "LESSON":
		return LabelableTypeLesson, nil
	case "USER_ASSET":
		return LabelableTypeUserAsset, nil
	default:
		var f LabelableType
		return f, fmt.Errorf("invalid LabelableType: %q", s)
	}
}

func (f LabelableType) String() string {
	switch f {
	case LabelableTypeComment:
		return "comment"
	case LabelableTypeLesson:
		return "lesson"
	case LabelableTypeUserAsset:
		return "user_asset"
	default:
		return "unknown"
	}
}
