package attr

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/rs/xid"
)

type Id struct {
	objType string
	value   string
}

func ParseId(id string) (*Id, error) {
	v, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return new(Id), fmt.Errorf("ParseId(%v): invalid id expected base64 encoded", id)
	}
	components := strings.Split(string(v), "_")
	if len(components) != 2 {
		return new(Id), fmt.Errorf("ParseId(%v): invalid id expected format TYPE_ID", id)
	}
	return &Id{objType: components[0], value: id}, nil
}

func NewId(objType string) *Id {
	id := xid.New()
	value := fmt.Sprintf("%v_%s", objType, id)
	value = base64.StdEncoding.EncodeToString([]byte(value))
	return &Id{objType: objType, value: value}
}

func (id *Id) Type() string {
	return id.objType
}

func (id *Id) String() string {
	return id.value
}
