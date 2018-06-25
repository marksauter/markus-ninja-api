package data

import (
	"time"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/sirupsen/logrus"
)

var accessLevelsWithFields = []perm.AccessLevel{
	perm.Create,
	perm.Read,
	perm.Update,
}
var accessLevelsWithoutFields = []perm.AccessLevel{
	perm.Connect,
	perm.Delete,
	perm.Disconnect,
}

type Permission struct {
	AccessLevel string      `db:"access_level"`
	Audience    string      `db:"audience"`
	CreatedAt   time.Time   `db:"created_at"`
	Id          string      `db:"id"`
	Field       pgtype.Text `db:"field"`
	Type        pgtype.Text `db:"type"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

type PermService struct {
	db Queryer
}

func NewPermService(db Queryer) *PermService {
	return &PermService{db}
}

const countPermByNodeTypeSQL = `
	SELECT COUNT(*)
	FROM permission
	WHERE type = $1
`

func (s *PermService) CountByNodeType(nodeType interface{}) (n int32, err error) {
	mType, err := perm.ParseNodeType(structs.Name(nodeType))
	if err != nil {
		return
	}
	mylog.Log.WithField("node_type", mType).Info("Perm.CountByNodeType(nodeType)")
	err = prepareQueryRow(
		s.db,
		"countPermByNodeType",
		countPermByNodeTypeSQL,
		mType,
	).Scan(&n)

	mylog.Log.WithField("n", n).Info("")
	return
}

// Creates a suite of permissions for the passed model.
//	- permissions for Create/Read/Update access for each field in model.
//  - permissions for Connect/Disconnect/Delete.
// Defaults audience to "AUTHENTICATED"
func (s *PermService) CreatePermissionSuite(model interface{}) error {
	mType, err := perm.ParseNodeType(structs.Name(model))
	if err != nil {
		return err
	}
	mylog.Log.WithField(
		"model",
		mType,
	).Info("CreatePermissionSuite(model)")

	fields := structs.Names(model)
	n := len(fields)*len(accessLevelsWithFields) + len(accessLevelsWithoutFields)
	permissions := make([][]interface{}, n)
	i := 0
	field := pgtype.Text{}
	for _, al := range accessLevelsWithFields {
		for _, f := range fields {
			id, _ := mytype.NewOID("Perm")
			field.Set(strcase.ToSnake(f))
			permissions[i] = []interface{}{
				id,
				al,
				mType,
				perm.Authenticated,
				field.String,
			}
			i += 1
		}
	}
	field.Set(nil)
	for _, al := range accessLevelsWithoutFields {
		id, _ := mytype.NewOID("Perm")
		permissions[i] = []interface{}{
			id,
			al,
			mType,
			perm.Authenticated,
			field.String,
		}
		i += 1
	}

	tx, err, newTx := beginTransaction(s.db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer rollbackTransaction(tx)
	}

	permSvc := NewPermService(tx)
	if err := permSvc.DeletePermissionSuite(model); err != nil {
		return err
	}

	copyCount, err := tx.CopyFrom(
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

	err = permSvc.UpdatePermissionSuite(model)
	if err != nil {
		return err
	}

	if newTx {
		err = commitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	return nil
}

const deletePermissionSuiteSQL = `
	DELETE FROM permission
	WHERE type = $1
`

func (s *PermService) DeletePermissionSuite(nodeType interface{}) error {
	mType, err := perm.ParseNodeType(structs.Name(nodeType))
	if err != nil {
		return err
	}
	mylog.Log.WithField(
		"node_type", mType,
	).Info("Perm.DeletePermissionSuite(node_type)")
	_, err = prepareExec(
		s.db,
		"deletePermissionSuite",
		deletePermissionSuiteSQL,
		mType,
	)
	if err != nil {
	}

	return nil
}

func (s *PermService) UpdatePermissionSuite(model interface{}) error {
	mType, err := perm.ParseNodeType(structs.Name(model))
	if err != nil {
		return err
	}
	mylog.Log.WithField(
		"model",
		mType,
	).Info("UpdatePermissionSuite(model)")
	permissableUserFields, err := perm.GetPermissableFields(model)
	if err != nil {
		mylog.Log.WithError(err).Fatal("failed to get user field permissions")
		return err
	}
	err = s.UpdateOperationForFields(
		perm.Operation{AccessLevel: perm.Read, NodeType: mType},
		permissableUserFields.Filter(perm.Read).Names(),
		perm.Everyone,
	)
	if err != nil {
		mylog.Log.WithError(err).Fatal("error during permission update")
		return err
	}

	err = s.UpdateOperationForFields(
		perm.Operation{AccessLevel: perm.Create, NodeType: mType},
		permissableUserFields.Filter(perm.Create).Names(),
		perm.Everyone,
	)
	if err != nil {
		mylog.Log.WithError(err).Fatal("error during permission update")
		return err
	}

	err = s.UpdateOperationForFields(
		perm.Operation{AccessLevel: perm.Update, NodeType: mType},
		permissableUserFields.Filter(perm.Update).Names(),
		perm.Everyone,
	)
	if err != nil {
		mylog.Log.WithError(err).Fatal("error during permission update")
		return err
	}

	return nil
}

const getPermissionByRoleNameSQL = `
		SELECT
			id,
			access_level,
			field,
			type,
			created_at,
			updated_at
		FROM
			permission p
		JOIN role_permission rp ON rp.permission_id = p.id
		JOIN role r ON r.name = rp.role
		WHERE r.name = $1
	`

func (s *PermService) GetByRoleName(
	roleName string,
) ([]Permission, error) {
	permissions := make([]Permission, 0)

	rows, err := s.db.Query(getPermissionByRoleNameSQL, roleName)
	if err != nil {
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		p := new(Permission)
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
	roles []string,
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
				FROM role_permission
				WHERE role = ANY($3)
			)
		)
	`
	groupBySQL := `
		Group By operation
	`
	var row *pgx.Row
	if len(roles) != 0 {
		row = s.db.QueryRow(
			permissionSQL+andRoleNameSQL+groupBySQL,
			o.AccessLevel,
			o.NodeType,
			roles,
		)
	} else {
		row = s.db.QueryRow(
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

const updatePermissionSQL = `
	UPDATE permission
	SET audience = $1
	WHERE id = $2
`

func (s *PermService) Update(permissionId string, a perm.Audience) error {
	_, err := s.db.Exec(updatePermissionSQL, a, permissionId)
	if err != nil {
		mylog.Log.WithError(err).Error("error during execution")
		return err
	}

	return nil
}

const updateOperationForFieldsSQL = `
	UPDATE permission
	SET audience = $1
	WHERE access_level = $2
		AND audience != $1
		AND type = $3
		AND field = ANY($4)
`

func (s *PermService) UpdateOperationForFields(
	o perm.Operation,
	fields []string,
	a perm.Audience,
) error {
	if len(fields) == 0 {
		return nil
	}
	_, err := s.db.Exec(
		updateOperationForFieldsSQL,
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
