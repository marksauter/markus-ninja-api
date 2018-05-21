package perm

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/fatih/structs"
)

type AccessLevel int

const (
	Read AccessLevel = iota
	Create
	Connect
	Delete
	Disconnect
	Update
)

func (al AccessLevel) String() string {
	switch al {
	case Read:
		return "Read"
	case Create:
		return "Create"
	case Connect:
		return "Connect"
	case Delete:
		return "Delete"
	case Disconnect:
		return "Disconnect"
	case Update:
		return "Update"
	default:
		return "unknown"
	}
}

func ParseAccessLevel(lvl string) (AccessLevel, error) {
	switch strings.ToLower(lvl) {
	case "read":
		return Read, nil
	case "create":
		return Create, nil
	case "connect":
		return Connect, nil
	case "delete":
		return Delete, nil
	case "disconnect":
		return Disconnect, nil
	case "update":
		return Update, nil
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

type Audience int

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

type NodeType int

const (
	EmailType NodeType = iota
	LessonType
	LessonCommentType
	StudyType
	UserType
	UserEmailType
)

func (nt NodeType) String() string {
	switch nt {
	case EmailType:
		return "Email"
	case LessonType:
		return "Lesson"
	case LessonCommentType:
		return "LessonComment"
	case StudyType:
		return "Study"
	case UserType:
		return "User"
	case UserEmailType:
		return "UserEmail"
	default:
		return "unknown"
	}
}

func ParseNodeType(nodeType string) (NodeType, error) {
	switch strings.ToLower(nodeType) {
	case "lesson":
		return LessonType, nil
	case "lesson_comment":
		return LessonCommentType, nil
	case "study":
		return StudyType, nil
	case "user":
		return UserType, nil
	case "user_email":
		return UserEmailType, nil
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
	case []byte:
		*nt, err = ParseNodeType(string(v))
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
	CreateEmail     = Operation{Create, EmailType}
	DeleteEmail     = Operation{Delete, EmailType}
	ReadEmail       = Operation{Read, EmailType}
	UpdateEmail     = Operation{Update, EmailType}
	ConnectEmail    = Operation{Connect, EmailType}
	DisconnectEmail = Operation{Disconnect, EmailType}

	CreateLesson     = Operation{Create, LessonType}
	DeleteLesson     = Operation{Delete, LessonType}
	ReadLesson       = Operation{Read, LessonType}
	UpdateLesson     = Operation{Update, LessonType}
	ConnectLesson    = Operation{Connect, LessonType}
	DisconnectLesson = Operation{Disconnect, LessonType}

	CreateLessonComment     = Operation{Create, LessonCommentType}
	DeleteLessonComment     = Operation{Delete, LessonCommentType}
	ReadLessonComment       = Operation{Read, LessonCommentType}
	UpdateLessonComment     = Operation{Update, LessonCommentType}
	ConnectLessonComment    = Operation{Connect, LessonCommentType}
	DisconnectLessonComment = Operation{Disconnect, LessonCommentType}

	CreateStudy     = Operation{Create, StudyType}
	DeleteStudy     = Operation{Delete, StudyType}
	ReadStudy       = Operation{Read, StudyType}
	UpdateStudy     = Operation{Update, StudyType}
	ConnectStudy    = Operation{Connect, StudyType}
	DisconnectStudy = Operation{Disconnect, StudyType}

	CreateUser     = Operation{Create, UserType}
	DeleteUser     = Operation{Delete, UserType}
	ReadUser       = Operation{Read, UserType}
	UpdateUser     = Operation{Update, UserType}
	ConnectUser    = Operation{Connect, UserType}
	DisconnectUser = Operation{Disconnect, UserType}
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
