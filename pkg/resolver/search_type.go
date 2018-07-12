package resolver

import (
	"fmt"
	"strings"
)

type SearchType int

const (
	SearchTypeLesson SearchType = iota
	SearchTypeStudy
	SearchTypeTopic
	SearchTypeUser
	SearchTypeUserAsset
)

func ParseSearchType(s string) (SearchType, error) {
	switch strings.ToUpper(s) {
	case "LESSON":
		return SearchTypeLesson, nil
	case "STUDY":
		return SearchTypeStudy, nil
	case "TOPIC":
		return SearchTypeTopic, nil
	case "USER":
		return SearchTypeUser, nil
	case "USER_ASSET":
		return SearchTypeUserAsset, nil
	default:
		var f SearchType
		return f, fmt.Errorf("invalid SearchType: %q", s)
	}
}

func (f SearchType) String() string {
	switch f {
	case SearchTypeLesson:
		return "lesson"
	case SearchTypeStudy:
		return "study"
	case SearchTypeTopic:
		return "topic"
	case SearchTypeUser:
		return "user"
	case SearchTypeUserAsset:
		return "user_asset"
	default:
		return "unknown"
	}
}
