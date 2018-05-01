package data

import (
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
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

type PermissionModel struct {
	AccessLevel string      `db:"access_level"`
	Audience    string      `db:"audience"`
	CreatedAt   time.Time   `db:"created_at"`
	Id          string      `db:"id"`
	Field       pgtype.Text `db:"field"`
	Type        string      `db:"type"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

type PermService struct {
	*mydb.DB
}

func NewPermService(db *mydb.DB) *PermService {
	return &PermService{db}
}

// Creates a suite of permissions for the passed node.
//	- permissions for Create/Read/Update access for each field in node.
//  - permissions for Connect/Disconnect/Delete.
// Defaults audience to "AUTHENTICATED"
func (s *PermService) CreatePermissionSuite(node interface{}) error {
	mType, err := perm.ParseNodeType(structs.Name(node))
	mylog.Log.WithField(
		"node",
		mType,
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
				oid.New("Perm"),
				al,
				mType,
				perm.Authenticated,
				field.Get(),
			}
			i += 1
		}
	}
	field.Set(nil)
	for _, al := range accessLevelsWithoutFields {
		permissions[i] = []interface{}{
			oid.New("Perm"),
			al,
			mType,
			perm.Authenticated,
			field.Get(),
		}
		i += 1
	}

	copyCount, err := s.CopyFrom(
		pgx.Identifier{"permission"},
		[]string{"id", "access_level", "type", "audience", "field"},
		pgx.CopyFromRows(permissions),
	)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			default:
				return err
			case UniqueViolation:
				mylog.Log.Warn("permissions already created")
				return nil
			}
		}
		return err
	}

	mylog.Log.Infof("created %v permissions for type %s", copyCount, mType)
	return nil
}

func (s *PermService) GetByRoleName(
	roleName string,
) ([]PermissionModel, error) {
	permissions := make([]PermissionModel, 0)

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
	rows, err := s.Query(permissionSQL, roleName)
	if err != nil {
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		p := new(PermissionModel)
		err := rows.Scan(
			&p.Id,
			&p.AccessLevel,
			&p.Field,
			&p.Type,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error during scan")
			return nil, err
		}
		permissions = append(permissions, *p)
	}

	if err = rows.Err(); err != nil {
		mylog.Log.WithError(err).Error("error during rows processing")
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
		"operation": o,
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
		AND (p.audience = 'EVERYONE'
			OR p.id IN (
				SELECT permission_id
				FROM role_permission rp
				INNER JOIN role r ON r.id = rp.role_id
				WHERE r.name = ANY($3)
			)
		)
	`
	groupBySQL := `
		Group By operation
	`
	var row *pgx.Row
	if len(roles) != 0 {
		row = s.QueryRow(
			permissionSQL+andRoleNameSQL+groupBySQL,
			o.AccessLevel,
			o.NodeType,
			roles,
		)
	} else {
		row = s.QueryRow(
			permissionSQL+andAudienceSQL+groupBySQL,
			o.AccessLevel,
			o.NodeType,
		)
	}
	var operation pgtype.Text
	var fields pgtype.TextArray
	err := row.Scan(&operation, &fields)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("error during scan")
		return nil, err
	}
	p.Operation, err = perm.ParseOperation(operation.String)
	if err != nil {
		mylog.Log.WithError(err).Error("error during parse operation")
		return nil, err
	}
	p.Fields = make([]string, len(fields.Elements))
	for i, f := range fields.Elements {
		p.Fields[i] = f.String
	}

	mylog.Log.WithField("fields", p.Fields).Info("query granted permission")
	return p, nil
}

func (s *PermService) Update(permissionId string, a perm.Audience) error {
	permissionSQL := `
		UPDATE permission
		SET audience = $1
		WHERE id = $2
	`
	_, err := s.Exec(permissionSQL, a, permissionId)
	if err != nil {
		mylog.Log.WithError(err).Error("error during execution")
		return err
	}

	return nil
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
			AND audience != $1
			AND type = $3
			AND field = ANY($4)
	`
	_, err := s.Exec(
		permissionSQL,
		a,
		o.AccessLevel,
		o.NodeType,
		fields,
	)
	if err != nil {
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"operation": o,
		"fields":    fields,
	}).Info("made permissions public")
	return nil
}
