package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type EnrollmentTypeValue int

const (
	EnrollmentTypeLesson EnrollmentTypeValue = iota
	EnrollmentTypeStudy
	EnrollmentTypeUser
)

func (f EnrollmentTypeValue) String() string {
	switch f {
	case EnrollmentTypeLesson:
		return "Lesson"
	case EnrollmentTypeStudy:
		return "Study"
	case EnrollmentTypeUser:
		return "User"
	default:
		return "unknown"
	}
}

type EnrollmentType struct {
	Status pgtype.Status
	V      EnrollmentTypeValue
}

func NewEnrollmentType(v EnrollmentTypeValue) EnrollmentType {
	return EnrollmentType{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseEnrollmentType(s string) (EnrollmentType, error) {
	switch strings.Title(s) {
	case "Lesson":
		return EnrollmentType{
			Status: pgtype.Present,
			V:      EnrollmentTypeLesson,
		}, nil
	case "Study":
		return EnrollmentType{
			Status: pgtype.Present,
			V:      EnrollmentTypeStudy,
		}, nil
	case "User":
		return EnrollmentType{
			Status: pgtype.Present,
			V:      EnrollmentTypeUser,
		}, nil
	default:
		var f EnrollmentType
		return f, fmt.Errorf("invalid EnrollmentType: %q", s)
	}
}

func (src *EnrollmentType) String() string {
	return src.V.String()
}

func (dst *EnrollmentType) Set(src interface{}) error {
	if src == nil {
		*dst = EnrollmentType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case EnrollmentType:
		*dst = value
		dst.Status = pgtype.Present
	case *EnrollmentType:
		*dst = *value
		dst.Status = pgtype.Present
	case EnrollmentTypeValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *EnrollmentTypeValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseEnrollmentType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseEnrollmentType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseEnrollmentType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to EnrollmentType", value)
	}

	return nil
}

func (src *EnrollmentType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *EnrollmentType) AssignTo(dst interface{}) error {
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

func (dst *EnrollmentType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = EnrollmentType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseEnrollmentType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *EnrollmentType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *EnrollmentType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *EnrollmentType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *EnrollmentType) Scan(src interface{}) error {
	if src == nil {
		*dst = EnrollmentType{Status: pgtype.Null}
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
func (src *EnrollmentType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
