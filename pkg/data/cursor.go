package data

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jackc/pgx"
)

type Cursor struct {
	String string
	Value  string
}

func NewCursor(cursor string) (*Cursor, error) {
	v, err := DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}
	c := &Cursor{
		String: cursor,
		Value:  v,
	}
	return c, nil
}

func (c *Cursor) SQL(field string, args *pgx.QueryArgs) string {
	return field + " = " + args.Append(c.Value)
}

func DecodeCursor(cursor string) (string, error) {
	bs, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	decodedValue := strings.TrimPrefix(string(bs), "cursor:")
	return decodedValue, nil
}

func EncodeCursor(value interface{}) (string, error) {
	cursor := ""
	switch v := value.(type) {
	case int32:
		cursor = fmt.Sprintf("cursor:%d", v)
	case string:
		cursor = fmt.Sprintf("cursor:%s", v)
	default:
		return "", fmt.Errorf("invalid type %T for value", v)
	}
	return base64.StdEncoding.EncodeToString([]byte(cursor)), nil
}
