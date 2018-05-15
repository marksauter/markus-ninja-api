package data

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Cursor struct {
	s *string
	v *string
}

func NewCursor(cursor *string) (*Cursor, error) {
	v, err := DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}
	c := &Cursor{
		s: cursor,
		v: v,
	}
	return c, nil
}

func (c *Cursor) String() string {
	if c.s != nil {
		return *c.s
	}
	return ""
}

func (c *Cursor) Value() string {
	if c.v != nil {
		return *c.v
	}
	return ""
}

func DecodeCursor(cursor *string) (*string, error) {
	var decodedValue string
	if cursor != nil {
		bs, err := base64.StdEncoding.DecodeString(*cursor)
		if err != nil {
			return nil, err
		}
		decodedValue = strings.TrimPrefix(string(bs), "cursor:")
	}
	return &decodedValue, nil
}

func EncodeCursor(id string) string {
	return base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("cursor:%s", id)),
	)
}
