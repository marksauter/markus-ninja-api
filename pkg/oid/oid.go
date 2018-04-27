package oid

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"
)

type ID struct {
	objType string
	value   string
}

var ErrInvalidID = errors.New("invalid ID")

func New(objType string) *ID {
	id := xid.New()
	value := fmt.Sprintf("%v_%s", objType, id)
	value = base64.StdEncoding.EncodeToString([]byte(value))
	return &ID{objType: objType, value: value}
}

func (id *ID) Type() string {
	return id.objType
}

func (id *ID) String() string {
	return id.value
}

func Parse(id string) (*ID, error) {
	v, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return new(ID), ErrInvalidID
	}
	components := strings.Split(string(v), "_")
	if len(components) != 2 {
		return new(ID), ErrInvalidID
	}
	return &ID{objType: components[0], value: id}, nil
}

func (o ID) Value() (driver.Value, error) {
	return o.String(), nil
}
