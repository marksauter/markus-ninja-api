package resolver

import (
	"fmt"
	"strings"
)

type SearchType int

const (
	SearchTypeActivity SearchType = iota
	SearchTypeCourse
	SearchTypeLabel
	SearchTypeLesson
	SearchTypeStudy
	SearchTypeTopic
	SearchTypeUser
	SearchTypeUserAsset
)

func ParseSearchType(s string) (SearchType, error) {
	switch strings.ToUpper(s) {
	case "ACTIVITY":
		return SearchTypeActivity, nil
	case "COURSE":
		return SearchTypeCourse, nil
	case "LABEL":
		return SearchTypeLabel, nil
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
	case SearchTypeActivity:
		return "activity"
	case SearchTypeCourse:
		return "course"
	case SearchTypeLabel:
		return "label"
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
