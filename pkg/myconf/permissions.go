package myconf

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type Permission struct {
	Operation     mytype.Operation `json:"operation"`
	Authenticated bool             `json:"authenticated"`
	Roles         []string         `json:"roles,omitempty,flow"`
	Fields        []string         `json:"fields,omitempty,flow"`
}

type Permissions struct {
	Permissions []Permission `json:"permissions,flow"`
}

func LoadPermissions() (*Permissions, error) {
	data, err := ioutil.ReadFile("./permissions.yml")
	if err != nil {
		return nil, err
	}

	permissions := &Permissions{}
	if err := yaml.Unmarshal(data, permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}
