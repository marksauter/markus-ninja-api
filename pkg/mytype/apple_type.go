package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type AppleTypeValue int

const (
	AppleTypeCourse AppleTypeValue = iota
	AppleTypeStudy
)

func (f AppleTypeValue) String() string {
	switch f {
	case AppleTypeCourse:
		return "Course"
	case AppleTypeStudy:
		return "Study"
	default:
		return "unknown"
	}
}

type AppleType struct {
	Status pgtype.Status
	V      AppleTypeValue
}

func NewAppleType(v AppleTypeValue) AppleType {
	return AppleType{
		Status: pgtype.Present,
		V:      v,
	}
}

func ParseAppleType(s string) (AppleType, error) {
	switch strings.Title(s) {
	case "Course":
		return AppleType{
			Status: pgtype.Present,
			V:      AppleTypeCourse,
		}, nil
	case "Study":
		return AppleType{
			Status: pgtype.Present,
			V:      AppleTypeStudy,
		}, nil
	default:
		var f AppleType
		return f, fmt.Errorf("invalid AppleType: %q", s)
	}
}

func (src *AppleType) String() string {
	return src.V.String()
}

func (dst *AppleType) Set(src interface{}) error {
	if src == nil {
		*dst = AppleType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case AppleType:
		*dst = value
		dst.Status = pgtype.Present
	case *AppleType:
		*dst = *value
		dst.Status = pgtype.Present
	case AppleTypeValue:
		dst.V = value
		dst.Status = pgtype.Present
	case *AppleTypeValue:
		dst.V = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseAppleType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseAppleType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseAppleType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to AppleType", value)
	}

	return nil
}

func (src *AppleType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *AppleType) AssignTo(dst interface{}) error {
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

func (dst *AppleType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = AppleType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseAppleType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *AppleType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *AppleType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, ErrUndefined
	}

	return append(buf, src.V.String()...), nil
}

func (src *AppleType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *AppleType) Scan(src interface{}) error {
	if src == nil {
		*dst = AppleType{Status: pgtype.Null}
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
func (src *AppleType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.V.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, ErrUndefined
	}
}
