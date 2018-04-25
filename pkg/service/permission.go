package service

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/sirupsen/logrus"
)

var accessLevelsWithFields = []perm.AccessLevel{
	perm.CreateAccess,
	perm.ReadAccess,
	perm.UpdateAccess,
}
var accessLevelsWithoutFields = []perm.AccessLevel{
	perm.ConnectAccess,
	perm.DeleteAccess,
	perm.DisconnectAccess,
}

type PermService struct {
	db *mydb.DB
}

func NewPermService(db *mydb.DB) *PermService {
	return &PermService{db: db}
}

func (s *PermService) Create(
	accessLevel perm.AccessLevel,
	audience perm.Audience,
	mType perm.NodeType,
	field string,
) (*model.Permission, error) {
	return nil, nil
}

// Creates a suite of permissions for the passed node.
//	- permissions for Create/Read/Update access for each field in node.
//  - permissions for Connect/Disconnect/Delete.
// Defaults audience to "AUTHENTICATED"
func (s *PermService) CreatePermissionSuite(node interface{}) error {
	mType, err := perm.ParseNodeType(structs.Name(node))
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
				attr.NewId("Perm").String(),
				al.String(),
				mType.String(),
				perm.Authenticated.String(),
				field.Get(),
			}
			i += 1
		}
	}
	field.Set(nil)
	for _, al := range accessLevelsWithoutFields {
		permissions[i] = []interface{}{
			attr.NewId("Perm").String(),
			al.String(),
			mType.String(),
			perm.Authenticated.String(),
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

func (s *PermService) GetByRoleName(
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

func (s *PermService) GetQueryPermission(
	o perm.Operation,
	roles ...string,
) (*perm.QueryPermission, error) {
	mylog.Log.WithFields(logrus.Fields{
		"operation": o.String(),
		"roles":     roles,
	}).Info("GetQueryPermission(operation, roles)")
	p := new(perm.QueryPermission)

	permissionSQL := `
		SELECT
			p.access_level || ' ' || p.type operation,
			array_agg(p.field) fields
		FROM
			permission p
		WHERE p.access_level = $1
			AND p.type = $2
	`
	andAudienceSQL := `
		AND p.audience = 'EVERYONE'
	`
	andRoleNameSQL := `
		AND p.id IN (
			SELECT permission_id
			FROM role_permission rp
			INNER JOIN role r ON r.id = rp.role_id
			WHERE r.name = ANY($3)
		)
	`
	groupBySQL := `
		Group By operation
	`
	var row *pgx.Row
	if len(roles) != 0 {
		row = s.db.QueryRow(
			permissionSQL+andRoleNameSQL+groupBySQL,
			o.AccessLevel.String(),
			o.NodeType.String(),
			roles,
		)
	} else {
		row = s.db.QueryRow(
			permissionSQL+andAudienceSQL+groupBySQL,
			o.AccessLevel.String(),
			o.NodeType.String(),
		)
	}
	var operation string
	err := row.Scan(
		&operation,
		&p.Fields,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error during scan")
		return nil, err
	}
	p.Operation, err = perm.ParseOperation(operation)
	if err != nil {
		mylog.Log.WithError(err).Error("error during parse operation")
		return nil, err
	}

	mylog.Log.Debug("query permission found")
	return p, nil
}

func (s *PermService) Update(permissionId string, a perm.Audience) (*model.Permission, error) {
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

func (s *PermService) UpdateOperationForFields(
	o perm.Operation,
	fields []string,
	a perm.Audience,
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
		o.AccessLevel.String(),
		o.NodeType.String(),
		fields,
	)
	if err != nil {
		return err
	}

	mylog.Log.Infof("made %v permissions public for fields %v", o, fields)
	return nil
}
