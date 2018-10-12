package mytype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type WordName struct {
	Status pgtype.Status
	String string
}

var validWordNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]{1,39}$`)

var InvalidWordNameLength = errors.New("name must be less than 40 characters")
var InvalidWordNameCharacters = errors.New("name may only contain alphanumeric characters or hyphens")

func checkWordName(name []byte) error {
	if len(name) > 39 {
		return InvalidWordNameLength
	}
	if match := validWordNameRegex.Match(name); !match {
		return InvalidWordNameCharacters
	}
	return nil
}

func (dst *WordName) Set(src interface{}) error {
	if src == nil {
		*dst = WordName{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case WordName:
		*dst = value
	case *WordName:
		*dst = *value
	case string:
		if err := checkWordName([]byte(value)); err != nil {
			return err
		}
		*dst = WordName{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = WordName{Status: pgtype.Null}
		} else {
			if err := checkWordName([]byte(*value)); err != nil {
				return err
			}
			*dst = WordName{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = WordName{Status: pgtype.Null}
		} else {
			if err := checkWordName(value); err != nil {
				return err
			}
			*dst = WordName{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to WordName", value)
	}

	return nil
}

func (dst *WordName) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *WordName) AssignTo(dst interface{}) error {
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

func (dst *WordName) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = WordName{Status: pgtype.Null}
		return nil
	}

	*dst = WordName{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *WordName) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *WordName) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *WordName) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *WordName) Scan(src interface{}) error {
	if src == nil {
		*dst = WordName{Status: pgtype.Null}
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
func (src *WordName) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
