package perm

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/fatih/structs"
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

func (al *AccessLevel) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*al, err = ParseAccessLevel(v)
		return
	default:
		err = fmt.Errorf("invalid type for access level %T", v)
		return
	}
}
func (al AccessLevel) Value() (driver.Value, error) {
	return al.String(), nil
}

type Audience int64

const (
	NoAudience Audience = iota
	Authenticated
	Everyone
)

func (a Audience) String() string {
	switch a {
	case Authenticated:
		return "AUTHENTICATED"
	case Everyone:
		return "EVERYONE"
	default:
		return "NOAUDIENCE"
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

func (a *Audience) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*a, err = ParseAudience(v)
		return
	default:
		return fmt.Errorf("invalid type for audience %T", v)
	}
}

func (a Audience) Value() (driver.Value, error) {
	return a.String(), nil
}

type NodeType int64

const (
	UserType NodeType = iota
	LessonType
)

func (nt NodeType) String() string {
	switch nt {
	case UserType:
		return "User"
	case LessonType:
		return "Lesson"
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

func (nt *NodeType) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*nt, err = ParseNodeType(v)
		return
	default:
		err = fmt.Errorf("invalid type for node type %T", v)
		return
	}
}

func (nt NodeType) Value() (driver.Value, error) {
	return nt.String(), nil
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

	CreateLesson     = Operation{CreateAccess, LessonType}
	DeleteLesson     = Operation{DeleteAccess, LessonType}
	ReadLesson       = Operation{ReadAccess, LessonType}
	UpdateLesson     = Operation{UpdateAccess, LessonType}
	ConnectLesson    = Operation{ConnectAccess, LessonType}
	DisconnectLesson = Operation{DisconnectAccess, LessonType}
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

func (o *Operation) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case string:
		*o, err = ParseOperation(v)
		return
	default:
		err = fmt.Errorf("invalid type for operation %T", v)
		return
	}
}

func (o Operation) Value() (driver.Value, error) {
	return o.String(), nil
}

type QueryPermission struct {
	Operation Operation
	Audience  Audience
	Fields    []string
}

func GetPermissableFields(model interface{}) (PermissableFields, error) {
	fields := structs.Fields(model)
	permissableFields := make([]*PermissableField, 0, len(fields))

	for _, field := range fields {
		permissableField, err := NewPermissableField(field)
		if err != nil {
			return nil, err
		}
		permissableFields = append(permissableFields, permissableField)
	}

	return PermissableFields(permissableFields), nil
}

type PermissableFields []*PermissableField

func (fs PermissableFields) Filter(al AccessLevel) PermissableFields {
	permissableFields := make([]*PermissableField, 0, len(fs))
	for _, f := range fs {
		if f.Can(al) {
			permissableFields = append(permissableFields, f)
		}
	}
	return PermissableFields(permissableFields)
}

func (fs PermissableFields) Names() []string {
	names := make([]string, len(fs))
	for i, f := range fs {
		names[i] = f.Name
	}
	return names
}

func NewPermissableField(f *structs.Field) (*PermissableField, error) {
	permissableField := &PermissableField{
		Name: f.Tag("db"),
	}
	permit := f.Tag("permit")
	if permit != "" {
		permissions := strings.Split(permit, "/")
		accessLookup := make(map[AccessLevel]bool, len(permissions))
		for _, p := range permissions {
			al, err := ParseAccessLevel(p)
			if err != nil {
				return nil, err
			}
			accessLookup[al] = true
		}
		permissableField.accessLookup = accessLookup
	}

	return permissableField, nil
}

type PermissableField struct {
	Name         string
	accessLookup map[AccessLevel]bool
}

func (fp *PermissableField) Can(al AccessLevel) bool {
	return fp.accessLookup[al]
}
