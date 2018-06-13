package mytype

import (
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type Filename struct {
	Status pgtype.Status
	String string
}

var invalidFilenameChar = regexp.MustCompile(`[^\w\.]+`)

const filenameReplacementStr = "-"

func (dst *Filename) Set(src interface{}) error {
	if src == nil {
		*dst = Filename{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Filename:
		*dst = value
	case *Filename:
		*dst = *value
	case string:
		value = invalidFilenameChar.ReplaceAllString(value, filenameReplacementStr)
		*dst = Filename{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Filename{Status: pgtype.Null}
		} else {
			*value = invalidFilenameChar.ReplaceAllString(*value, filenameReplacementStr)
			*dst = Filename{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Filename{Status: pgtype.Null}
		} else {
			value = invalidFilenameChar.ReplaceAll(value, []byte(filenameReplacementStr))
			*dst = Filename{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to Filename", value)
	}

	return nil
}

func (dst *Filename) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Filename) AssignTo(dst interface{}) error {
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

func (dst *Filename) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Filename{Status: pgtype.Null}
		return nil
	}

	*dst = Filename{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *Filename) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Filename) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *Filename) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Filename) Scan(src interface{}) error {
	if src == nil {
		*dst = Filename{Status: pgtype.Null}
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
func (src *Filename) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
