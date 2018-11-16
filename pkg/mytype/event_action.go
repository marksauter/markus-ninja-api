package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type EventActionValue int

const (
	CreatedAction EventActionValue = iota
	AddedToCourseAction
	RemovedFromCourseAction
	AppledAction
	UnappledAction
	CommentedAction
	LabeledAction
	UnlabeledAction
	MentionedAction
	PublishedAction
	ReferencedAction
	RenamedAction
)

func (f EventActionValue) String() string {
	switch f {
	case CreatedAction:
		return "created"
	case AddedToCourseAction:
		return "added_to_course"
	case RemovedFromCourseAction:
		return "removed_from_course"
	case AppledAction:
		return "appled"
	case UnappledAction:
		return "unappled"
	case CommentedAction:
		return "commented"
	case LabeledAction:
		return "labeled"
	case UnlabeledAction:
		return "unlabeled"
	case MentionedAction:
		return "mentioned"
	case PublishedAction:
		return "published"
	case ReferencedAction:
		return "referenced"
	case RenamedAction:
		return "renamed"
	default:
		return "unknown"
	}
}

type EventAction struct {
	Status pgtype.Status
	V      EventActionValue
}

func NewEventAction(v EventActionValue) EventAction {
	return EventAction{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseEventAction(s string) (EventAction, error) {
	switch strings.Title(s) {
	case "created":
		return EventAction{
			Status: pgtype.Present,
			V:      CreatedAction,
		}, nil
	case "added_to_course":
		return EventAction{
			Status: pgtype.Present,
			V:      AddedToCourseAction,
		}, nil
	case "removed_from_course":
		return EventAction{
			Status: pgtype.Present,
			V:      RemovedFromCourseAction,
		}, nil
	case "appled":
		return EventAction{
			Status: pgtype.Present,
			V:      AppledAction,
		}, nil
	case "unappled":
		return EventAction{
			Status: pgtype.Present,
			V:      UnappledAction,
		}, nil
	case "commented":
		return EventAction{
			Status: pgtype.Present,
			V:      CommentedAction,
		}, nil
	case "labeled":
		return EventAction{
			Status: pgtype.Present,
			V:      LabeledAction,
		}, nil
	case "unlabeled":
		return EventAction{
			Status: pgtype.Present,
			V:      UnlabeledAction,
		}, nil
	case "mentioned":
		return EventAction{
			Status: pgtype.Present,
			V:      MentionedAction,
		}, nil
	case "published":
		return EventAction{
			Status: pgtype.Present,
			V:      PublishedAction,
		}, nil
	case "referenced":
		return EventAction{
			Status: pgtype.Present,
			V:      ReferencedAction,
		}, nil
	case "renamed":
		return EventAction{
			Status: pgtype.Present,
			V:      RenamedAction,
		}, nil
	default:
		var f EventAction
		return f, fmt.Errorf("invalid EventAction: %q", s)
	}
}

func (src *EventAction) String() string {
	return src.V.String()
}

func (dst *EventAction) Set(src interface{}) error {
	if src == nil {
		*dst = EventAction{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case EventAction:
		*dst = value
		dst.Status = pgtype.Present
	case *EventAction:
		*dst = *value
		dst.Status = pgtype.Present
	case EventActionValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *EventActionValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseEventAction(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseEventAction(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseEventAction(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to EventAction", value)
	}

	return nil
}

func (src *EventAction) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *EventAction) AssignTo(dst interface{}) error {
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

func (dst *EventAction) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = EventAction{Status: pgtype.Null}
		return nil
	}

	t, err := ParseEventAction(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *EventAction) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *EventAction) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *EventAction) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *EventAction) Scan(src interface{}) error {
	if src == nil {
		*dst = EventAction{Status: pgtype.Null}
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
func (src *EventAction) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
