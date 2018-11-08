package mytype

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordEmpty = errors.New("password cannot be empty")
var ErrPasswordTooWeak = errors.New("password too weak")

type PasswordStrength int

const (
	VeryWeak PasswordStrength = iota
	Weak
	Moderate
	Strong
	VeryStrong
)

type Password struct {
	Bytes  []byte
	Status pgtype.Status
	value  string
}

func NewPassword(password string) (*Password, error) {
	if password == "" {
		return nil, ErrPasswordEmpty
	}
	return &Password{
		Bytes:  hash(password),
		Status: pgtype.Present,
		value:  password,
	}, nil
}

func hash(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// This should never happen, so panic if it does
		panic(err)
	}
	return hash
}

func (p *Password) CheckStrength(s PasswordStrength) error {
	entropy := zxcvbn.PasswordStrength(p.value, nil)
	strength := PasswordStrength(entropy.Score)
	mylog.Log.Debug(strength)
	if strength < s {
		return ErrPasswordTooWeak
	}
	return nil
}

func (p *Password) CompareToPassword(password string) error {
	return bcrypt.CompareHashAndPassword(p.Bytes, []byte(password))
}

func (dst *Password) Set(src interface{}) error {
	if src == nil {
		*dst = Password{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case string:
		pwrd, err := NewPassword(value)
		if err != nil {
			return err
		}
		*dst = *pwrd
	case *string:
		pwrd, err := NewPassword(*value)
		if err != nil {
			return err
		}
		*dst = *pwrd
	case []byte:
		if value != nil {
			pwrd, err := NewPassword(string(value))
			if err != nil {
				return err
			}
			*dst = *pwrd
		} else {
			*dst = Password{Status: pgtype.Null}
		}
	default:
		return fmt.Errorf("cannot convert %v to Password", value)
	}

	return nil
}

func (dst *Password) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.Bytes
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Password) AssignTo(dst interface{}) error {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *[]byte:
			buf := make([]byte, len(src.Bytes))
			copy(buf, src.Bytes)
			*v = buf
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

// DecodeText only supports the hex format. This has been the default since
// PostgreSQL 9.0.
func (dst *Password) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Password{Status: pgtype.Null}
		return nil
	}

	if len(src) < 2 || src[0] != '\\' || src[1] != 'x' {
		return fmt.Errorf("invalid hex format")
	}

	buf := make([]byte, (len(src)-2)/2)
	_, err := hex.Decode(buf, src[2:])
	if err != nil {
		return err
	}

	*dst = Password{Bytes: buf, Status: pgtype.Present}
	return nil
}

func (dst *Password) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = Password{Status: pgtype.Null}
		return nil
	}

	*dst = Password{Bytes: src, Status: pgtype.Present}
	return nil
}

func (src *Password) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	buf = append(buf, `\x`...)
	buf = append(buf, hex.EncodeToString(src.Bytes)...)
	return buf, nil
}

func (src *Password) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.Bytes...), nil
}

// Scan implements the database/sql Scanner interface.
func (dst *Password) Scan(src interface{}) error {
	if src == nil {
		*dst = Password{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		buf := make([]byte, len(src))
		copy(buf, src)
		*dst = Password{Bytes: buf, Status: pgtype.Present}
		return nil
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src *Password) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.Bytes, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
