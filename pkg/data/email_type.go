package data

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/pgtype"
)

type EmailTypeValue int

const (
	BackupEmail EmailTypeValue = iota
	ExtraEmail
	PrimaryEmail
)

func (src EmailTypeValue) String() string {
	switch src {
	case BackupEmail:
		return "BACKUP"
	case ExtraEmail:
		return "EXTRA"
	case PrimaryEmail:
		return "PRIMARY"
	default:
		return "unknown"
	}
}

type EmailType struct {
	Status pgtype.Status
	Type   EmailTypeValue
}

func NewEmailType(v EmailTypeValue) EmailType {
	return EmailType{
		Status: pgtype.Present,
		Type:   v,
	}
}

func ParseEmailType(s string) (EmailType, error) {
	switch strings.ToUpper(s) {
	case "BACKUP":
		return EmailType{
			Status: pgtype.Present,
			Type:   BackupEmail,
		}, nil
	case "EXTRA":
		return EmailType{
			Status: pgtype.Present,
			Type:   ExtraEmail,
		}, nil
	case "PRIMARY":
		return EmailType{
			Status: pgtype.Present,
			Type:   PrimaryEmail,
		}, nil
	default:
		var o EmailType
		return o, fmt.Errorf("invalid EmailType: %q", s)
	}
}

func (src *EmailType) String() string {
	return src.Type.String()
}

func (dst *EmailType) Set(src interface{}) error {
	if src == nil {
		*dst = EmailType{Status: pgtype.Null}
	}
	switch value := src.(type) {
	case EmailType:
		*dst = value
		dst.Status = pgtype.Present
	case *EmailType:
		*dst = *value
		dst.Status = pgtype.Present
	case EmailTypeValue:
		dst.Type = value
		dst.Status = pgtype.Present
	case *EmailTypeValue:
		dst.Type = *value
		dst.Status = pgtype.Present
	case string:
		t, err := ParseEmailType(value)
		if err != nil {
			return err
		}
		*dst = t
	case *string:
		t, err := ParseEmailType(*value)
		if err != nil {
			return err
		}
		*dst = t
	case []byte:
		t, err := ParseEmailType(string(value))
		if err != nil {
			return err
		}
		*dst = t
	default:
		return fmt.Errorf("cannot convert %v to EmailType", value)
	}

	return nil
}

func (src *EmailType) Get() interface{} {
	switch src.Status {
	case pgtype.Present:
		return src
	case pgtype.Null:
		return nil
	default:
		return src.Status
	}
}

func (src *EmailType) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *string:
			*v = src.Type.String()
			return nil
		case *[]byte:
			*v = make([]byte, len(src.Type.String()))
			copy(*v, src.Type.String())
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

func (dst *EmailType) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = EmailType{Status: pgtype.Null}
		return nil
	}

	t, err := ParseEmailType(string(src))
	if err != nil {
		return err
	}
	*dst = t
	return nil
}

func (dst *EmailType) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

var errUndefined = errors.New("cannot encode status undefined")

func (src *EmailType) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.Type.String()...), nil
}

func (src *EmailType) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *EmailType) Scan(src interface{}) error {
	if src == nil {
		*dst = EmailType{Status: pgtype.Null}
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
func (src *EmailType) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.Type.String(), nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
