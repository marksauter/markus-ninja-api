package service

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type AccessLevel int64
type Audience string
type ModelType string

const (
	ReadAccess AccessLevel = iota
	CreateAccess
	ConnectAccess
	DeleteAccess
	DisconnectAccess
	UpdateAccess

	Authenticated Audience = "AUTHENTICATED"
	Everyone      Audience = "EVERYONE"

	UserType ModelType = "User"
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
		return al, fmt.Errorf("not a valid access level: %q", lvl)
	}
}

func (al *AccessLevel) Scan(value interface{}) error {
	*al = AccessLevel(value.(int64))
	return nil
}
func (al AccessLevel) Value() (driver.Value, error) {
	return int64(al), nil
}

func (a *Audience) Scan(value interface{}) error {
	*a = Audience(value.(string))
	return nil
}
func (a Audience) Value() (driver.Value, error) {
	return string(a), nil
}

func (mt *ModelType) Scan(value interface{}) error {
	*mt = ModelType(value.(string))
	return nil
}
func (mt ModelType) Value() (driver.Value, error) {
	return string(mt), nil
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
	mType ModelType,
	field string,
) (*model.Permission, error) {
	return nil, nil
}

func (s *PermissionService) CreatePermissionSuite(model interface{}) error {
	mName := structs.Name(model)
	mylog.Log.WithField("model", mName).Info("CreatePermissionSuite(model)")

	// mType := ModelType(mName)
	fields := structs.Names(model)
	n := len(fields)*len(accessLevelsWithFields) + len(accessLevelsWithoutFields)
	permissions := make([][]interface{}, n)
	i := 0
	field := pgtype.Text{}
	for _, al := range accessLevelsWithFields {
		for _, f := range fields {
			field.Set(f)
			permissions[i] = []interface{}{
				attr.NewId("Permission").String(),
				al.String(),
				mName,
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
			mName,
			field.Get(),
		}
		i += 1
	}

	copyCount, err := s.db.CopyFrom(
		pgx.Identifier{"permission"},
		[]string{"id", "access_level", "type", "field"},
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

	mylog.Log.Infof("created %v permissions for type %v", copyCount, mName)
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
			&p.ID,
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
