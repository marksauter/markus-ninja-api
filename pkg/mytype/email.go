package mytype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/pgtype"
)

type Email struct {
	Status pgtype.Status
	String string
}

var ErrNoMatch = errors.New("name string did not pass validation")

var validateEmail = regexp.MustCompile(
	`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`,
)

func (dst *Email) Set(src interface{}) error {
	if src == nil {
		*dst = Email{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case Email:
		*dst = value
	case *Email:
		*dst = *value
	case string:
		if ok := validateEmail.MatchString(value); !ok {
			return ErrNoMatch
		}
		*dst = Email{String: value, Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = Email{Status: pgtype.Null}
		} else {
			if ok := validateEmail.MatchString(*value); !ok {
				return ErrNoMatch
			}
			*dst = Email{String: *value, Status: pgtype.Present}
		}
	case []byte:
		if value == nil {
			*dst = Email{Status: pgtype.Null}
		} else {
			if ok := validateEmail.Match(value); !ok {
				return ErrNoMatch
			}
			*dst = Email{String: string(value), Status: pgtype.Present}
		}
	default:
		return fmt.Errorf("cannot convert %v to Email", value)
	}

	return nil
}

func (dst *Email) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Email) AssignTo(dst interface{}) error {
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

func (dst *Email) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Email{Status: pgtype.Null}
		return nil
	}

	*dst = Email{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *Email) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *Email) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *Email) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *Email) Scan(src interface{}) error {
	if src == nil {
		*dst = Email{Status: pgtype.Null}
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
func (src *Email) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
