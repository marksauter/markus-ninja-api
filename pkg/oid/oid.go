package oid

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/xid"
)

// Object ID
type OID struct {
	// Unique part of the OID without the type information.
	Short string
	// Type of object for the OID.
	Type string
	// Base64 encoded value of the OID.
	value string
}

var ErrInvalidID = errors.New("invalid OID")

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
	value := fmt.Sprintf("%03d%s%s", n, objType, id)
	value = base64.StdEncoding.EncodeToString([]byte(value))
	return &OID{Short: id.String(), Type: objType, value: value}, nil
}

func (o OID) String() string {
	return o.value
}

func Parse(id string) (*OID, error) {
	v, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	s := string(v)
	nStr := s[:3]
	n, err := strconv.ParseInt(nStr, 10, 16)
	if err != nil {
		return nil, ErrInvalidID
	}
	t := s[4 : 4+n]
	short := s[4+n:]
	return &OID{Short: short, Type: t, value: id}, nil
}

func (o OID) Value() (driver.Value, error) {
	return o.value, nil
}
