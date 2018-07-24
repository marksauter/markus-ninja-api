package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type EnrollmentStatusValue int

const (
	EnrollmentStatusEnrolled EnrollmentStatusValue = iota
	EnrollmentStatusIgnored
	EnrollmentStatusUnenrolled
)

func (f EnrollmentStatusValue) String() string {
	switch f {
	case EnrollmentStatusEnrolled:
		return "ENROLLED"
	case EnrollmentStatusIgnored:
		return "IGNORED"
	case EnrollmentStatusUnenrolled:
		return "UNENROLLED"
	default:
		return "unknown"
	}
}

type EnrollmentStatus struct {
	Status pgtype.Status
	V      EnrollmentStatusValue
}

func NewEnrollmentStatus(v EnrollmentStatusValue) EnrollmentStatus {
	return EnrollmentStatus{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseEnrollmentStatus(s string) (EnrollmentStatus, error) {
	switch strings.ToUpper(s) {
	case "ENROLLED":
		return EnrollmentStatus{
			Status: pgtype.Present,
			V:      EnrollmentStatusEnrolled,
		}, nil
	case "IGNORED":
		return EnrollmentStatus{
			Status: pgtype.Present,
			V:      EnrollmentStatusIgnored,
		}, nil
	case "UNENROLLED":
		return EnrollmentStatus{
			Status: pgtype.Present,
			V:      EnrollmentStatusUnenrolled,
		}, nil
	default:
		var f EnrollmentStatus
		return f, fmt.Errorf("invalid EnrollmentStatus: %q", s)
	}
}

func (src *EnrollmentStatus) String() string {
	return src.V.String()
}

func (dst *EnrollmentStatus) Set(src interface{}) error {
	if src == nil {
		*dst = EnrollmentStatus{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case EnrollmentStatus:
		*dst = value
		dst.Status = pgtype.Present
	case *EnrollmentStatus:
		*dst = *value
		dst.Status = pgtype.Present
	case EnrollmentStatusValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *EnrollmentStatusValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseEnrollmentStatus(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseEnrollmentStatus(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseEnrollmentStatus(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to EnrollmentStatus", value)
	}

	return nil
}

func (src *EnrollmentStatus) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *EnrollmentStatus) AssignTo(dst interface{}) error {
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

func (dst *EnrollmentStatus) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = EnrollmentStatus{Status: pgtype.Null}
		return nil
	}

	t, err := ParseEnrollmentStatus(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *EnrollmentStatus) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *EnrollmentStatus) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *EnrollmentStatus) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *EnrollmentStatus) Scan(src interface{}) error {
	if src == nil {
		*dst = EnrollmentStatus{Status: pgtype.Null}
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
func (src *EnrollmentStatus) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
