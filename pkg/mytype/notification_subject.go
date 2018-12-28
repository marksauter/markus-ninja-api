package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type NotificationSubjectValue int

const (
	NotificationSubjectLesson NotificationSubjectValue = iota
	NotificationSubjectUserAsset
)

func (f NotificationSubjectValue) String() string {
	switch f {
	case NotificationSubjectLesson:
		return "Lesson"
	case NotificationSubjectUserAsset:
		return "UserAsset"
	default:
		return "unknown"
	}
}

type NotificationSubject struct {
	Status pgtype.Status
	V      NotificationSubjectValue
}

func NewNotificationSubject(v NotificationSubjectValue) NotificationSubject {
	return NotificationSubject{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseNotificationSubject(s string) (NotificationSubject, error) {
	switch strings.Title(s) {
	case "Lesson":
		return NotificationSubject{
			Status: pgtype.Present,
			V:      NotificationSubjectLesson,
		}, nil
	case "UserAsset":
		return NotificationSubject{
			Status: pgtype.Present,
			V:      NotificationSubjectUserAsset,
		}, nil
	default:
		var f NotificationSubject
		return f, fmt.Errorf("invalid NotificationSubject: %q", s)
	}
}

func (src *NotificationSubject) String() string {
	return src.V.String()
}

func (dst *NotificationSubject) Set(src interface{}) error {
	if src == nil {
		*dst = NotificationSubject{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case NotificationSubject:
		*dst = value
		dst.Status = pgtype.Present
	case *NotificationSubject:
		*dst = *value
		dst.Status = pgtype.Present
	case NotificationSubjectValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *NotificationSubjectValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseNotificationSubject(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseNotificationSubject(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseNotificationSubject(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to NotificationSubject", value)
	}

	return nil
}

func (src *NotificationSubject) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *NotificationSubject) AssignTo(dst interface{}) error {
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

func (dst *NotificationSubject) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = NotificationSubject{Status: pgtype.Null}
		return nil
	}

	t, err := ParseNotificationSubject(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *NotificationSubject) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *NotificationSubject) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *NotificationSubject) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *NotificationSubject) Scan(src interface{}) error {
	if src == nil {
		*dst = NotificationSubject{Status: pgtype.Null}
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
func (src *NotificationSubject) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
