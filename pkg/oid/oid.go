package oid

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func New(objType string) (*OID, error) {
	if objType == "" {
		return nil, errors.New("invalid OID: `objType` must not be empty")
	}
	id := xid.New()
	objType = strings.Title(strings.ToLower(objType))
	n := len(objType)
	if n > 999 {
		return nil, errors.New("invalid OID: `objType` too long")
	}
	s := fmt.Sprintf("%03d%s%s", n, objType, id)
	s = base64.StdEncoding.EncodeToString([]byte(s))
	return &OID{Short: id.String(), Type: objType, String: s}, nil
}

var errInvalidID = errors.New("invalid OID")

func Parse(id string) (*OID, error) {
	v, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, errInvalidID
	}
	s := string(v)
	nStr := s[:3]
	n, err := strconv.ParseInt(nStr, 10, 16)
	if err != nil {
		return nil, errInvalidID
	}
	t := s[4 : 4+n]
	short := s[4+n:]
	return &OID{Short: short, Status: pgtype.Present, Type: t, String: id}, nil
}

func (dst *OID) Set(src interface{}) (err error) {
	if src == nil {
		*dst = OID{Status: pgtype.Null}
		return nil
	}

	switch value := src.(type) {
	case string:
		dst, err = Parse(value)
		return
	case *string:
		dst, err = Parse(*value)
		return
	case []byte:
		dst, err = Parse(string(value))
		return
	default:
		return fmt.Errorf("cannot convert %v to OID", value)
	}
}

func (dst *OID) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst
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

	*dst = OID{String: string(src), Status: pgtype.Present}
	return nil
}

func (dst *OID) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	return dst.DecodeText(ci, src)
}
