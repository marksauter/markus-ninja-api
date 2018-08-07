package resolver

import (
	"fmt"
	"strings"
)

type TopicableType int

const (
	TopicableTypeStudy TopicableType = iota
	TopicableTypeCourse
)

func ParseTopicableType(s string) (TopicableType, error) {
	switch strings.ToUpper(s) {
	case "COURSE":
		return TopicableTypeCourse, nil
	case "STUDY":
		return TopicableTypeStudy, nil
	default:
		var f TopicableType
		return f, fmt.Errorf("invalid TopicableType: %q", s)
	}
}

func (f TopicableType) String() string {
	switch f {
	case TopicableTypeCourse:
		return "course"
	case TopicableTypeStudy:
		return "study"
	default:
		return "unknown"
	}
}
