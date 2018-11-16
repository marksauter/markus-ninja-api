package mytype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type WordsName struct {
	Status pgtype.Status
	String string
}

var validWordsNameRegex = regexp.MustCompile(`^[\w-]{1,39}$`)

var InvalidWordsNameLength = errors.New("name must be at least one character but less than 40 characters")
var InvalidWordsNameCharacters = errors.New("name may only contain alphanumeric characters, hyphens, or underscores")

func checkWordsName(name []byte) error {
	if len(name) < 1 || len(name) > 39 {
		return InvalidWordsNameLength
	}
	if match := validWordsNameRegex.Match(name); !match {
		return InvalidWordsNameCharacters
	}
	return nil
}

func (dst *WordsName) Set(src interface{}) error {
	if src == nil {
		*dst = WordsName{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case WordsName:
		*dst = value
	case *WordsName:
		*dst = *value
	case string:
		if err := checkWordsName([]byte(value)); err != nil {
			return err
		}
		*dst = WordsName{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = WordsName{Status: pgtype.Null}
		} else {
			if err := checkWordsName([]byte(*value)); err != nil {
				return err
			}
			*dst = WordsName{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = WordsName{Status: pgtype.Null}
		} else {
			if err := checkWordsName(value); err != nil {
				return err
			}
			*dst = WordsName{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to WordsName", value)
	}

	return nil
}

func (dst *WordsName) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *WordsName) AssignTo(dst interface{}) error {
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

func (dst *WordsName) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = WordsName{Status: pgtype.Null}
		return nil
	}

	*dst = WordsName{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *WordsName) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *WordsName) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *WordsName) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *WordsName) Scan(src interface{}) error {
	if src == nil {
		*dst = WordsName{Status: pgtype.Null}
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
func (src *WordsName) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
