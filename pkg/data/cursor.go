package data

import (
	"encoding/base64"
	"fmt"
	"strings"
)

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
