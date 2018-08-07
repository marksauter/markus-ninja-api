package mytype

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Color struct {
	Status pgtype.Status
	String string
}

var InvalidColorError = errors.New("color must be in hexadecimal format")

func (dst *Color) Set(src interface{}) error {
	if src == nil {
		*dst = Color{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Color:
		*dst = value
	case *Color:
		*dst = *value
	case string:
		if !util.IsHexColor(value) {
			return InvalidColorError
		}
		*dst = Color{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Color{Status: pgtype.Null}
		} else {
			if !util.IsHexColor(*value) {
				return InvalidColorError
			}
			*dst = Color{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Color{Status: pgtype.Null}
		} else {
			if !util.IsHexColor(string(value)) {
				return InvalidColorError
			}
			*dst = Color{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to Color", value)
	}

	return nil
}

func (dst *Color) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Color) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *string:
			*v = src.String
			return nil
		case *[]byte:
			*v = make([]byte, len(src.String))
			copy(*v, src.String)
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

func (dst *Color) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Color{Status: pgtype.Null}
		return nil
	}

	*dst = Color{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *Color) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Color) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *Color) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Color) Scan(src interface{}) error {
	if src == nil {
		*dst = Color{Status: pgtype.Null}
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
func (src *Color) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
