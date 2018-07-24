package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type NodeType int

const (
	AppledNodeType NodeType = iota
	EmailNodeType
	EnrolledNodeType
	EventNodeType
	EVTNodeType
	LabelNodeType
	LabeledNodeType
	LessonNodeType
	LessonCommentNodeType
	NotificationNodeType
	PRTNodeType
	StudyNodeType
	TopicNodeType
	TopicedNodeType
	UserNodeType
	UserAssetNodeType
)

func (nt NodeType) String() string {
	switch nt {
	case AppledNodeType:
		return "Appled"
	case EmailNodeType:
		return "Email"
	case EnrolledNodeType:
		return "Enrolled"
	case EventNodeType:
		return "Event"
	case EVTNodeType:
		return "EVT"
	case LabelNodeType:
		return "Label"
	case LabeledNodeType:
		return "Labeled"
	case LessonNodeType:
		return "Lesson"
	case LessonCommentNodeType:
		return "LessonComment"
	case NotificationNodeType:
		return "Notification"
	case PRTNodeType:
		return "PRT"
	case StudyNodeType:
		return "Study"
	case TopicNodeType:
		return "Topic"
	case TopicedNodeType:
		return "Topiced"
	case UserNodeType:
		return "User"
	case UserAssetNodeType:
		return "UserAsset"
	default:
		return "unknown"
	}
}

func ParseNodeType(nodeType string) (NodeType, error) {
	switch strings.ToLower(nodeType) {
	case "appled":
		return AppledNodeType, nil
	case "email":
		return EmailNodeType, nil
	case "enrolled":
		return EnrolledNodeType, nil
	case "event":
		return EventNodeType, nil
	case "evt":
		return EVTNodeType, nil
	case "label":
		return LabelNodeType, nil
	case "labeled":
		return LabeledNodeType, nil
	case "lesson":
		return LessonNodeType, nil
	case "lessoncomment":
		return LessonCommentNodeType, nil
	case "notification":
		return NotificationNodeType, nil
	case "prt":
		return PRTNodeType, nil
	case "study":
		return StudyNodeType, nil
	case "topic":
		return TopicNodeType, nil
	case "topiced":
		return TopicedNodeType, nil
	case "user":
		return UserNodeType, nil
	case "userasset":
		return UserAssetNodeType, nil
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
