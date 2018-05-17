package oid

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/maybe"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/rs/xid"
)

// Object ID
type OID struct {
	// Unique part of the OID without the type information.
	Short string
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
	return &OID{Short: short, Type: t, String: id}, nil
}

type MaybeOID struct {
	Status maybe.MaybeStatus
	oid    *OID
}

func NewMaybe(objType string) (*MaybeOID, error) {
	oid, err := New(objType)
	if err != nil {
		return nil, err
	}
	return &MaybeOID{Status: maybe.Just, oid: oid}, nil
}

func (src *MaybeOID) Get() interface{} {
	switch src.Status {
	case maybe.Just:
		return src.oid
	default:
		return src.Status
	}
}

func (dst *MaybeOID) Just(src interface{}) error {
	if src == nil {
		*dst = MaybeOID{Status: maybe.Nothing}
		return nil
	}

	switch value := src.(type) {
	case OID:
		*dst = MaybeOID{Status: maybe.Just, oid: &value}
	case *OID:
		*dst = MaybeOID{Status: maybe.Just, oid: value}
	default:
		return fmt.Errorf("cannot convert %v to MaybeOID", value)
	}
	return nil
}

// func (dst *MaybeOID) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
//   mylog.Log.Debug("DecodeText")
//   if src == nil {
//     *dst = MaybeOID{Status: maybe.Just}
//     return nil
//   }
//
//   oid, err := Parse(string(src))
//   if err != nil {
//     return err
//   }
//   *dst = MaybeOID{Status: maybe.Just, oid: oid}
//   return nil
// }
//
// func (dst *MaybeOID) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
//   mylog.Log.Debug("DecodeBinary")
//   return dst.DecodeText(ci, src)
// }

func (src *MaybeOID) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	mylog.Log.Debug("EncodeText")
	switch src.Status {
	case maybe.Just:
		// Treating empty string as null value in db.
		if src.oid.String == "" {
			return nil, nil
		}
		return append(buf, src.oid.String...), nil
	default:
		return nil, errNothing
	}
}

func (src *MaybeOID) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	mylog.Log.Debug("EncodeBinary")
	return src.EncodeText(ci, buf)
}

var errNothing = errors.New("cannot encode status nothing")

func (dst *MaybeOID) Scan(src interface{}) error {
	mylog.Log.Debug("Scan")
	if src == nil {
		*dst = MaybeOID{Status: maybe.Just}
		return nil
	}
	switch src := src.(type) {
	case string:
		oid, err := Parse(src)
		if err != nil {
			return err
		}
		*dst = MaybeOID{Status: maybe.Just, oid: oid}
		return nil
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		oid, err := Parse(string(srcCopy))
		if err != nil {
			return err
		}
		*dst = MaybeOID{Status: maybe.Just, oid: oid}
		return nil
	}

	return fmt.Errorf("cannot scan %T", src)
}

func (src *MaybeOID) Value() (driver.Value, error) {
	mylog.Log.Debug("Value")
	switch src.Status {
	case maybe.Just:
		// Treating empty string as null value in db.
		if src.oid.String == "" {
			return nil, nil
		}
		return src.oid.String, nil
	default:
		return nil, errNothing
	}
}
