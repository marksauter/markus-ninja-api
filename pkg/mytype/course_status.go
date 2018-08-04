package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type CourseStatusValue int

const (
	CourseStatusAdvancing CourseStatusValue = iota
	CourseStatusCompleted
)

func (f CourseStatusValue) String() string {
	switch f {
	case CourseStatusAdvancing:
		return "ADVANCING"
	case CourseStatusCompleted:
		return "COMPLETED"
	default:
		return "unknown"
	}
}

type CourseStatus struct {
	Status pgtype.Status
	V      CourseStatusValue
}

func NewCourseStatus(v CourseStatusValue) CourseStatus {
	return CourseStatus{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseCourseStatus(s string) (CourseStatus, error) {
	switch strings.ToUpper(s) {
	case "ADVANCING":
		return CourseStatus{
			Status: pgtype.Present,
			V:      CourseStatusAdvancing,
		}, nil
	case "COMPLETED":
		return CourseStatus{
			Status: pgtype.Present,
			V:      CourseStatusCompleted,
		}, nil
	default:
		var f CourseStatus
		return f, fmt.Errorf("invalid CourseStatus: %q", s)
	}
}

func (src *CourseStatus) String() string {
	return src.V.String()
}

func (dst *CourseStatus) Set(src interface{}) error {
	if src == nil {
		*dst = CourseStatus{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case CourseStatus:
		*dst = value
		dst.Status = pgtype.Present
	case *CourseStatus:
		*dst = *value
		dst.Status = pgtype.Present
	case CourseStatusValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *CourseStatusValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseCourseStatus(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseCourseStatus(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseCourseStatus(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to CourseStatus", value)
	}

	return nil
}

func (src *CourseStatus) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *CourseStatus) AssignTo(dst interface{}) error {
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

func (dst *CourseStatus) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = CourseStatus{Status: pgtype.Null}
		return nil
	}

	t, err := ParseCourseStatus(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *CourseStatus) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *CourseStatus) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *CourseStatus) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *CourseStatus) Scan(src interface{}) error {
	if src == nil {
		*dst = CourseStatus{Status: pgtype.Null}
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
func (src *CourseStatus) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
