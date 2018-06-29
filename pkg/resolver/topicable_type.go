package resolver

import (
	"fmt"
	"strings"
)

type TopicableType int

const (
	TopicableTypeStudy TopicableType = iota
)

func ParseTopicableType(s string) (TopicableType, error) {
	switch strings.ToUpper(s) {
	case "STUDY":
		return TopicableTypeStudy, nil
	default:
		var f TopicableType
		return f, fmt.Errorf("invalid TopicableType: %q", s)
	}
}

func (f TopicableType) String() string {
	switch f {
	case TopicableTypeStudy:
		return "study"
	default:
		return "unknown"
	}
}
