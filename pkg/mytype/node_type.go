package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type NodeType int

const (
	AppledType NodeType = iota
	EmailType
	EnrolledType
	EventType
	EVTType
	LabelType
	LabeledType
	LessonType
	LessonCommentType
	NotificationType
	PRTType
	RefType
	StudyType
	TopicType
	TopicedType
	UserType
	UserAssetType
)

func (nt NodeType) String() string {
	switch nt {
	case AppledType:
		return "Appled"
	case EmailType:
		return "Email"
	case EnrolledType:
		return "Enrolled"
	case EventType:
		return "Event"
	case EVTType:
		return "EVT"
	case LabelType:
		return "Label"
	case LabeledType:
		return "Labeled"
	case LessonType:
		return "Lesson"
	case LessonCommentType:
		return "LessonComment"
	case NotificationType:
		return "Notification"
	case PRTType:
		return "PRT"
	case RefType:
		return "Ref"
	case StudyType:
		return "Study"
	case TopicType:
		return "Topic"
	case TopicedType:
		return "Topiced"
	case UserType:
		return "User"
	case UserAssetType:
		return "UserAsset"
	default:
		return "unknown"
	}
}

func ParseNodeType(nodeType string) (NodeType, error) {
	switch strings.ToLower(nodeType) {
	case "appled":
		return AppledType, nil
	case "email":
		return EmailType, nil
	case "enrolled":
		return EnrolledType, nil
	case "event":
		return EventType, nil
	case "evt":
		return EVTType, nil
	case "label":
		return LabelType, nil
	case "labeled":
		return LabeledType, nil
	case "lesson":
		return LessonType, nil
	case "lessoncomment":
		return LessonCommentType, nil
	case "notification":
		return NotificationType, nil
	case "prt":
		return PRTType, nil
	case "ref":
		return RefType, nil
	case "study":
		return StudyType, nil
	case "topic":
		return TopicType, nil
	case "topiced":
		return TopicedType, nil
	case "user":
		return UserType, nil
	case "userasset":
		return UserAssetType, nil
	default:
		var t NodeType
		return t, fmt.Errorf("invalid node type: %q", nodeType)
	}
}

func (nt *NodeType) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*nt, err = ParseNodeType(v)
		return
	case []byte:
		*nt, err = ParseNodeType(string(v))
		return
	default:
		err = fmt.Errorf("invalid type for node type %T", v)
		return
	}
}

func (nt NodeType) Value() (driver.Value, error) {
	return nt.String(), nil
}
