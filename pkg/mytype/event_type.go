package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type EventTypeValue int

const (
	CourseEvent EventTypeValue = iota
	LessonEvent
	PublicEvent
	StudyEvent
	UserAssetEvent
)

func (f EventTypeValue) String() string {
	switch f {
	case CourseEvent:
		return "CourseEvent"
	case LessonEvent:
		return "LessonEvent"
	case PublicEvent:
		return "PublicEvent"
	case StudyEvent:
		return "StudyEvent"
	case UserAssetEvent:
		return "UserAssetEvent"
	default:
		return "unknown"
	}
}

type EventType struct {
	Status pgtype.Status
	V      EventTypeValue
}

func NewEventType(v EventTypeValue) EventType {
	return EventType{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseEventType(s string) (EventType, error) {
	switch strings.Title(s) {
	case "CourseEvent":
		return EventType{
			Status: pgtype.Present,
			V:      CourseEvent,
		}, nil
	case "LessonEvent":
		return EventType{
			Status: pgtype.Present,
			V:      LessonEvent,
		}, nil
	case "PublicEvent":
		return EventType{
			Status: pgtype.Present,
			V:      PublicEvent,
		}, nil
	case "StudyEvent":
		return EventType{
			Status: pgtype.Present,
			V:      StudyEvent,
		}, nil
	case "UserAssetEvent":
		return EventType{
			Status: pgtype.Present,
			V:      UserAssetEvent,
		}, nil
	default:
		var f EventType
		return f, fmt.Errorf("invalid EventType: %q", s)
	}
}

func (src *EventType) String() string {
	return src.V.String()
}

func (dst *EventType) Set(src interface{}) error {
	if src == nil {
		*dst = EventType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case EventType:
		*dst = value
		dst.Status = pgtype.Present
	case *EventType:
		*dst = *value
		dst.Status = pgtype.Present
	case EventTypeValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *EventTypeValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseEventType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseEventType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseEventType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to EventType", value)
	}

	return nil
}

func (src *EventType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *EventType) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *string:
			*v = src.V.String()
			return nil
		case *[]byte:
			*v = make([]byte, len(src.V.String()))
			copy(*v, src.V.String())
			return nil
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
				return src.AssignTo(nextDst)
			}
		}
	case pgtype.Null:
		return pgtype.NullAssignTo(dst)
	}

	return fmt.Errorf("cannot decode %v into %T", src, dst)
}

func (dst *EventType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = EventType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseEventType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *EventType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *EventType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *EventType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *EventType) Scan(src interface{}) error {
	if src == nil {
		*dst = EventType{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return dst.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src *EventType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
