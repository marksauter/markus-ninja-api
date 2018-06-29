package resolver

import (
	"fmt"
	"strings"
)

type AppleableType int

const (
	AppleableTypeStudy AppleableType = iota
)

func ParseAppleableType(s string) (AppleableType, error) {
	switch strings.ToUpper(s) {
	case "STUDY":
		return AppleableTypeStudy, nil
	default:
		var f AppleableType
		return f, fmt.Errorf("invalid AppleableType: %q", s)
	}
}

func (f AppleableType) String() string {
	switch f {
	case AppleableTypeStudy:
		return "study"
	default:
		return "unknown"
	}
}
