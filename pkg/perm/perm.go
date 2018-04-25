package perm

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type AccessLevel int64

const (
	ReadAccess AccessLevel = iota
	CreateAccess
	ConnectAccess
	DeleteAccess
	DisconnectAccess
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
		return al, fmt.Errorf("invalid access level: %q", lvl)
	}
}

func (al *AccessLevel) Scan(value interface{}) error {
	*al = AccessLevel(value.(int64))
	return nil
}
func (al AccessLevel) Value() (driver.Value, error) {
	return int64(al), nil
}

type Audience int64

const (
	Authenticated Audience = iota
	Everyone
)

func (a Audience) String() string {
	switch a {
	case Authenticated:
		return "AUTHENTICATED"
	case Everyone:
		return "EVERYONE"
	default:
		return "unknown"
	}
}

func ParseAudience(aud string) (Audience, error) {
	switch strings.ToLower(aud) {
	case "authenticated":
		return Authenticated, nil
	case "everyone":
		return Everyone, nil
	default:
		var a Audience
		return a, fmt.Errorf("invalid audience: %q", aud)
	}
}

func (a *Audience) Scan(value interface{}) error {
	*a = Audience(value.(int64))
	return nil
}
func (a Audience) Value() (driver.Value, error) {
	return int64(a), nil
}

type NodeType int64

const (
	UserType NodeType = iota
)

func (nt NodeType) String() string {
	switch nt {
	case UserType:
		return "User"
	default:
		return "unknown"
	}
}

func ParseNodeType(nodeType string) (NodeType, error) {
	switch strings.ToLower(nodeType) {
	case "user":
		return UserType, nil
	default:
		var t NodeType
		return t, fmt.Errorf("invalid node type: %q", nodeType)
	}
}

func (nt *NodeType) Scan(value interface{}) error {
	*nt = NodeType(value.(int64))
	return nil
}
func (nt NodeType) Value() (driver.Value, error) {
	return int64(nt), nil
}

type Operation struct {
	AccessLevel AccessLevel
	NodeType    NodeType
}

var (
	CreateUser     = Operation{CreateAccess, UserType}
	DeleteUser     = Operation{DeleteAccess, UserType}
	ReadUser       = Operation{ReadAccess, UserType}
	UpdateUser     = Operation{UpdateAccess, UserType}
	ConnectUser    = Operation{ConnectAccess, UserType}
	DisconnectUser = Operation{DisconnectAccess, UserType}
)

func (o Operation) String() string {
	return o.AccessLevel.String() + " " + o.NodeType.String()
}

func ParseOperation(operation string) (Operation, error) {
	var o Operation

	parsedOperation := strings.SplitN(strings.ToLower(operation), " ", 2)
	if len(parsedOperation) != 2 {
		return o, fmt.Errorf("invalid operation: %q", operation)
	}
	accessLevel, err := ParseAccessLevel(parsedOperation[0])
	if err != nil {
		return o, fmt.Errorf("invalid operation: %v", err)
	}
	o.AccessLevel = accessLevel
	nodeType, err := ParseNodeType(parsedOperation[1])
	if err != nil {
		return o, fmt.Errorf("invalid operation: %v", err)
	}
	o.NodeType = nodeType

	return o, nil
}

type QueryPermission struct {
	Operation Operation
	Fields    []string
}
