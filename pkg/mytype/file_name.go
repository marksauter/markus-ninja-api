package mytype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type Filename struct {
	Status pgtype.Status
	String string
}

var dotCheckRegex = regexp.MustCompile(`(^.|..|.$)`)
var validFilenameRegex = regexp.MustCompile(`^[\w-.]{1,39}$`)

var InvalidFilenameLength = errors.New("filename must be less than 40 characters")
var InvalidFilenameHyphens = errors.New("filename cannot have multiple consecutive dots, and cannot begin or end with a dot")
var InvalidFilenameCharacters = errors.New("filename may only contain alphanumeric characters, hyphens, underscores, or dots")

func checkFilename(filename []byte) error {
	if len(filename) > 39 {
		return InvalidFilenameLength
	}
	if match := hyphenCheckRegex.Match(filename); match {
		return InvalidFilenameHyphens
	}
	if match := validFilenameRegex.Match(filename); !match {
		return InvalidFilenameCharacters
	}
	return nil
}

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
		if err := checkFilename([]byte(value)); err != nil {
			return err
		}
		*dst = Filename{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Filename{Status: pgtype.Null}
		} else {
			if err := checkFilename([]byte(*value)); err != nil {
				return err
			}
			*dst = Filename{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Filename{Status: pgtype.Null}
		} else {
			if err := checkFilename(value); err != nil {
				return err
			}
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
