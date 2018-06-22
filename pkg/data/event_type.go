package data

type EventType int

const (
	CreateEvent EventType = iota
	DeleteEvent
	DismissEvent
	EnrollEvent
	MentionEvent
)

func (src EventType) String() string {
	switch src {
	case CreateEvent:
		return "created"
	case DeleteEvent:
		return "deleted"
	case DismissEvent:
		return "dismissed"
	case EnrollEvent:
		return "enrolled"
	case MentionEvent:
		return "mentioned"
	default:
		return "unknown"
	}
}
