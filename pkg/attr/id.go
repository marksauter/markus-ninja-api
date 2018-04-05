package attr

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Id struct {
	Type  string
	value string
}

func NewId(id string) (*Id, error) {
	idComponents := strings.Split(id, "_")
	if len(idComponents) != 2 {
		return new(Id), fmt.Errorf(`Invalid id "%v": expected format "Type_id"`, id)
	}
	return &Id{Type: idComponents[0], value: id}, nil
}

func (id *Id) String() string {
	return id.value
}

func (id *Id) Encode() string {
	return base64.StdEncoding.EncodeToString([]byte(id.value))
}
