package mytype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type Username struct {
	Status pgtype.Status
	String string
}

var hyphenCheckRegex = regexp.MustCompile(`(^-|--|-$)`)
var validUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]{1,39}$`)

var InvalidUsernameLength = errors.New("username must be less than 40 characters")
var InvalidUsernameHyphens = errors.New("username cannot have multiple consecutive hyphens, and cannot begin or end with a hyphen")
var InvalidUsernameCharacters = errors.New("username may only contain alphanumeric characters, hyphens, or underscores")

func checkUsername(username []byte) error {
	if len(username) > 39 {
		return InvalidUsernameLength
	}
	if match := hyphenCheckRegex.Match(username); match {
		return InvalidUsernameHyphens
	}
	if match := validUsernameRegex.Match(username); !match {
		return InvalidUsernameCharacters
	}
	return nil
}

func (dst *Username) Set(src interface{}) error {
	if src == nil {
		*dst = Username{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Username:
		*dst = value
	case *Username:
		*dst = *value
	case string:
		if err := checkUsername([]byte(value)); err != nil {
			return err
		}
		*dst = Username{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Username{Status: pgtype.Null}
		} else {
			if err := checkUsername([]byte(*value)); err != nil {
				return err
			}
			*dst = Username{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Username{Status: pgtype.Null}
		} else {
			if err := checkUsername(value); err != nil {
				return err
			}
			*dst = Username{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to Username", value)
	}

	return nil
}

func (dst *Username) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Username) AssignTo(dst interface{}) error {
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

func (dst *Username) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Username{Status: pgtype.Null}
		return nil
	}

	*dst = Username{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *Username) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Username) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *Username) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Username) Scan(src interface{}) error {
	if src == nil {
		*dst = Username{Status: pgtype.Null}
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
func (src *Username) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
