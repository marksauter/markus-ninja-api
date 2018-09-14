package resolver

import (
	"fmt"
	"strings"
)

type AppleableType int

const (
	AppleableTypeCourse AppleableType = iota
	AppleableTypeStudy
)

func ParseAppleableType(s string) (AppleableType, error) {
	switch strings.ToUpper(s) {
	case "COURSE":
		return AppleableTypeCourse, nil
	case "STUDY":
		return AppleableTypeStudy, nil
	default:
		var f AppleableType
		return f, fmt.Errorf("invalid AppleableType: %q", s)
	}
}

func (f AppleableType) String() string {
	switch f {
	case AppleableTypeCourse:
		return "course"
	case AppleableTypeStudy:
		return "study"
	default:
		return "unknown"
	}
}
