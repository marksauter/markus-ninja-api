package resolver

import (
	"fmt"
	"strings"
)

type EnrollableType int

const (
	EnrollableTypeLesson EnrollableType = iota
	EnrollableTypeStudy
	EnrollableTypeUser
)

func ParseEnrollableType(s string) (EnrollableType, error) {
	switch strings.ToUpper(s) {
	case "LESSON":
		return EnrollableTypeLesson, nil
	case "STUDY":
		return EnrollableTypeStudy, nil
	case "USER":
		return EnrollableTypeUser, nil
	default:
		var f EnrollableType
		return f, fmt.Errorf("invalid EnrollableType: %q", s)
	}
}

func (f EnrollableType) String() string {
	switch f {
	case EnrollableTypeLesson:
		return "lesson"
	case EnrollableTypeStudy:
		return "study"
	case EnrollableTypeUser:
		return "user"
	default:
		return "unknown"
	}
}
