package mytype

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type AccessLevel int

const (
	CreateAccess AccessLevel = iota
	ConnectAccess
	DeleteAccess
	DisconnectAccess
	ReadAccess
	UpdateAccess
)

func (al AccessLevel) String() string {
	switch al {
	case ReadAccess:
		return "Read"
	case CreateAccess:
		return "Create"
	case ConnectAccess:
		return "Connect"
	case DeleteAccess:
		return "Delete"
	case DisconnectAccess:
		return "Disconnect"
	case UpdateAccess:
		return "Update"
	default:
		return "unknown"
	}
}

func ParseAccessLevel(lvl string) (AccessLevel, error) {
	switch strings.ToLower(lvl) {
	case "read":
		return ReadAccess, nil
	case "create":
		return CreateAccess, nil
	case "connect":
		return ConnectAccess, nil
	case "delete":
		return DeleteAccess, nil
	case "disconnect":
		return DisconnectAccess, nil
	case "update":
		return UpdateAccess, nil
	default:
		var al AccessLevel
		return al, fmt.Errorf("invalid AccessLevel: %q", lvl)
	}
}

func (al *AccessLevel) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*al, err = ParseAccessLevel(v)
		return
	default:
		err = fmt.Errorf("invalid type for AccessLevel %T", v)
		return
	}
}

func (al AccessLevel) Value() (driver.Value, error) {
	return al.String(), nil
}
