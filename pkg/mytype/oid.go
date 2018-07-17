package mytype

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/pgtype"
	"github.com/rs/xid"
)

// Object ID
type OID struct {
	// Unique part of the OID without the type information.
	Short  string
	Status pgtype.Status
	// Base64 encoded value of the OID.
	String string
	// Type of object for the OID.
	Type string
}

func NewOID(objType string) (*OID, error) {
	if objType == "" {
		return nil, errors.New("invalid OID: `objType` must not be empty")
	}
	id := xid.New()
	n := len(objType)
	if n > 999 {
		return nil, errors.New("invalid OID: `objType` too long")
	}
	s := fmt.Sprintf("%03d%s%s", n, objType, id)
	s = base64.StdEncoding.EncodeToString([]byte(s))
	return &OID{Short: id.String(), Status: pgtype.Present, Type: objType, String: s}, nil
}

func NewOIDFromShort(objType, short string) (*OID, error) {
	if objType == "" {
		return nil, errors.New("invalid OID: `objType` must not be empty")
	}
	objType = strings.Title(strings.ToLower(objType))
	n := len(objType)
	if n > 999 {
		return nil, errors.New("invalid OID: `objType` too long")
	}
	s := fmt.Sprintf("%03d%s%s", n, objType, short)
	s = base64.StdEncoding.EncodeToString([]byte(s))
	return &OID{Short: short, Status: pgtype.Present, Type: objType, String: s}, nil
}

var errInvalidOID = errors.New("invalid OID")

func ParseOID(id string) (*OID, error) {
	if id == "" {
		return &OID{Status: pgtype.Null}, nil
	}
	v, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, errInvalidOID
	}
	s := string(v)
	if len(s) < 3 {
		return nil, errInvalidOID
	}
	nStr := s[:3]
	n, err := strconv.ParseInt(nStr, 10, 16)
	if err != nil {
		return nil, errInvalidOID
	}
	if len(s) < 3+int(n) {
		return nil, errInvalidOID
	}
	t := s[3 : 3+n]
	short := s[3+n:]
	return &OID{Short: short, Status: pgtype.Present, Type: t, String: id}, nil
}

func (src *OID) DBVarName() string {
	return fmt.Sprintf("%s_id", strcase.ToSnake(src.Type))
}

func (dst *OID) Set(src interface{}) error {
	if src == nil {
		*dst = OID{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case OID:
		*dst = value
	case *OID:
		*dst = *value
	case string:
		oid, err := ParseOID(value)
		if err != nil {
			return err
		}
		*dst = *oid
	case *string:
		if value == nil {
			*dst = OID{Status: pgtype.Null}
		} else {
			oid, err := ParseOID(*value)
			if err != nil {
				return err
			}
			*dst = *oid
		}
	case []byte:
		if value == nil {
			*dst = OID{Status: pgtype.Null}
		} else {
			oid, err := ParseOID(string(value))
			if err != nil {
				return err
			}
			*dst = *oid
		}
	default:
		return fmt.Errorf("cannot convert %v to OID", value)
	}

	return nil
}

func (dst *OID) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.String
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *OID) AssignTo(dst interface{}) error {
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

func (dst *OID) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = OID{Status: pgtype.Null}
		return nil
	}

	oid, err := ParseOID(string(src))
	if err != nil {
		return err
	}
	*dst = *oid
	return nil
}

func (dst *OID) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}

func (src *OID) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.String...), nil
}

func (src *OID) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	return src.EncodeText(ci, buf)
}

// Scan implements the database/sql Scanner interface.
func (dst *OID) Scan(src interface{}) error {
	if src == nil {
		*dst = OID{Status: pgtype.Null}
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
func (src *OID) Value() (driver.Value, error) {
	switch src.Status {
	case pgtype.Present:
		return src.String, nil
	case pgtype.Null:
		return nil, nil
	default:
		return nil, errUndefined
	}
}
