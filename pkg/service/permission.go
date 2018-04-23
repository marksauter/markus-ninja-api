package service

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
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
	accessLevel AccessLevel
	nodeType    NodeType
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
	return o.accessLevel.String() + " " + o.nodeType.String()
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
	o.accessLevel = accessLevel
	nodeType, err := ParseNodeType(parsedOperation[1])
	if err != nil {
		return o, fmt.Errorf("invalid operation: %v", err)
	}
	o.nodeType = nodeType

	return o, nil
}

var accessLevelsWithFields = []AccessLevel{
	CreateAccess,
	ReadAccess,
	UpdateAccess,
}
var accessLevelsWithoutFields = []AccessLevel{
	ConnectAccess,
	DeleteAccess,
	DisconnectAccess,
}

type PermissionService struct {
	db *mydb.DB
}

func NewPermissionService(db *mydb.DB) *PermissionService {
	return &PermissionService{db: db}
}

func (s *PermissionService) Create(
	accessLevel AccessLevel,
	audience Audience,
	mType NodeType,
	field string,
) (*model.Permission, error) {
	return nil, nil
}

// Creates a suite of permissions for the passed node.
//	- permissions for Create/Read/Update access for each field in node.
//  - permissions for Connect/Disconnect/Delete.
// Defaults audience to "AUTHENTICATED"
func (s *PermissionService) CreatePermissionSuite(node interface{}) error {
	mType, err := ParseNodeType(structs.Name(node))
	mylog.Log.WithField(
		"node",
		mType.String(),
	).Info("CreatePermissionSuite(node)")

	fields := structs.Names(node)
	n := len(fields)*len(accessLevelsWithFields) + len(accessLevelsWithoutFields)
	permissions := make([][]interface{}, n)
	i := 0
	field := pgtype.Text{}
	for _, al := range accessLevelsWithFields {
		for _, f := range fields {
			field.Set(strcase.ToSnake(f))
			permissions[i] = []interface{}{
				attr.NewId("Permission").String(),
				al.String(),
				mType.String(),
				Authenticated.String(),
				field.Get(),
			}
			i += 1
		}
	}
	field.Set(nil)
	for _, al := range accessLevelsWithoutFields {
		permissions[i] = []interface{}{
			attr.NewId("Permission").String(),
			al.String(),
			mType.String(),
			Authenticated.String(),
			field.Get(),
		}
		i += 1
	}

	copyCount, err := s.db.CopyFrom(
		pgx.Identifier{"permission"},
		[]string{"id", "access_level", "type", "audience", "field"},
		pgx.CopyFromRows(permissions),
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			switch mydb.PSQLError(pgErr.Code) {
			default:
				return err
			case mydb.UniqueViolation:
				mylog.Log.Warn("permissions already created")
				return nil
			}
		}
		return err
	}

	mylog.Log.Infof("created %v permissions for type %v", copyCount, mType.String())
	return nil
}

func (s *PermissionService) GetByRoleName(
	roleName string,
) ([]model.Permission, error) {
	permissions := make([]model.Permission, 0)

	permissionSQL := `
		SELECT
			id,
			access_level,
			field,
			type,
			created_at,
			updated_at
		FROM
			permission p
		INNER JOIN role_permission rp ON rp.permission_id = p.id
		INNER JOIN role r ON r.id = rp.role_id
		WHERE r.name = $1
	`
	rows, err := s.db.Query(permissionSQL, roleName)
	if err != nil {
		mylog.Log.WithField("error", err).Error("error during query")
		return nil, err
	}
	for i := 0; rows.Next(); i++ {
		p := new(model.Permission)
		err := rows.Scan(
			&p.Id,
			&p.AccessLevel,
			&p.Field,
			&p.Type,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			mylog.Log.WithField("error", err).Error("error during scan")
			return nil, err
		}
		permissions = append(permissions, *p)
	}

	if err = rows.Err(); err != nil {
		mylog.Log.WithField("error", err).Error("error during rows processing")
		return nil, err
	}

	mylog.Log.Debug("permissions found")
	return permissions, nil
}

func (s *PermissionService) Update(permissionId string, a Audience) (*model.Permission, error) {
	permissionSQL := `
		UPDATE permission
		SET audience = $1
		WHERE id = $2
		RETURNING
			id,
			access_level,
			audience,
			field,
			type,
			created_at,
			updated_at
	`
	p := new(model.Permission)
	row := s.db.QueryRow(permissionSQL, a.String(), permissionId)
	err := row.Scan(
		&p.Id,
		&p.AccessLevel,
		&p.Audience,
		&p.Field,
		&p.Type,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return p, fmt.Errorf(`permissions with id "%v" not found`)
		default:
			mylog.Log.WithError(err).Error("error during scan")
			return nil, err
		}
	}

	return p, nil
}

func (s *PermissionService) UpdateOperationForFields(
	o Operation,
	fields []string,
	a Audience,
) error {
	permissionSQL := `
		UPDATE permission
		SET audience = $1
		WHERE access_level = $2
			AND type = $3
			AND field = ANY($4)
	`
	_, err := s.db.Exec(
		permissionSQL,
		a.String(),
		o.accessLevel.String(),
		o.nodeType.String(),
		fields,
	)
	if err != nil {
		return err
	}

	mylog.Log.Infof("made %v permissions public for fields %v", o, fields)
	return nil
}
