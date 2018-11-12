package data

import (
	"errors"
	"strings"

	"github.com/fatih/structs"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
	"github.com/sirupsen/logrus"
)

var accessLevelsWithFields = []mytype.AccessLevel{
	mytype.CreateAccess,
	mytype.ReadAccess,
	mytype.UpdateAccess,
}
var accessLevelsWithoutFields = []mytype.AccessLevel{
	mytype.ConnectAccess,
	mytype.DeleteAccess,
	mytype.DisconnectAccess,
}

type Permission struct {
	AccessLevel pgtype.Text        `db:"access_level"`
	Audience    pgtype.Text        `db:"audience"`
	CreatedAt   pgtype.Timestamptz `db:"created_at"`
	GrantedAt   pgtype.Timestamptz `db:"granted_at"`
	ID          mytype.OID         `db:"id"`
	Field       pgtype.Text        `db:"field"`
	Type        pgtype.Text        `db:"type"`
	UpdatedAt   pgtype.Timestamptz `db:"updated_at"`
}

const countPermissionByTypeSQL = `
	SELECT COUNT(*)
	FROM permission
	WHERE type = $1
`

func CountPermissionByType(
	db Queryer,
	nodeType interface{},
) (int32, error) {
	mylog.Log.Info("CountPermissionByType()")
	var n int32
	mType, err := mytype.ParseNodeType(structs.Name(nodeType))
	if err != nil {
		return n, err
	}
	err = prepareQueryRow(
		db,
		"countPermissionByType",
		countPermissionByTypeSQL,
		mType,
	).Scan(&n)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
	} else {
		mylog.Log.WithField("n", n).Info(util.Trace("permissions found"))
	}
	return n, err
}

// Creates a suite of permissions for the passed model.
//	- permissions for Create/Read/Update access for each field in model.
//  - permissions for Connect/Disconnect/Delete.
// Defaults audience to "AUTHENTICATED"
func CreatePermissionSuite(
	db Queryer,
	model interface{},
) error {
	mylog.Log.Info("CreatePermissionSuite()")

	mType, err := mytype.ParseNodeType(structs.Name(model))
	if err != nil {
		return err
	}

	fields := structs.Fields(model)
	n := len(fields)*len(accessLevelsWithFields) + len(accessLevelsWithoutFields)
	permissions := make([][]interface{}, 0, n)
	for _, al := range accessLevelsWithFields {
		for _, f := range fields {
			id, _ := mytype.NewOID("Permission")
			field := f.Tag("db")
			permits := strings.Split(f.Tag("permit"), "/")
			for _, p := range permits {
				if strings.ToLower(p) == strings.ToLower(al.String()) {
					permissions = append(permissions, []interface{}{
						id,
						al,
						mType,
						mytype.Authenticated,
						field,
					})
				}
			}
		}
	}
	for _, al := range accessLevelsWithoutFields {
		id, _ := mytype.NewOID("Permission")
		permissions = append(permissions, []interface{}{
			id,
			al,
			mType,
			mytype.Authenticated,
			nil,
		})
	}

	tx, err, newTx := BeginTransaction(db)
	if err != nil {
		mylog.Log.WithError(err).Error("error starting transaction")
		return err
	}
	if newTx {
		defer RollbackTransaction(tx)
	}

	if err := DeletePermissionSuite(db, model); err != nil {
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

	if newTx {
		err = CommitTransaction(tx)
		if err != nil {
			mylog.Log.WithError(err).Error("error during transaction")
			return err
		}
	}

	mylog.Log.Infof("created %v permissions for type %s", copyCount, mType)

	return nil
}

func ConnectRolePermissions(
	db Queryer,
	o *mytype.Operation,
	fields,
	roles []string,
) error {
	mylog.Log.Info("ConnectRolePermissions()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	andFieldsSQL := ""
	if len(fields) != 0 {
		for i, f := range fields {
			fields[i] = strings.ToLower(f)
		}
		andFieldsSQL = `AND permission.field = ANY(` + args.Append(fields) + `)`
	}
	joinRolesOnSQL := "true"
	if len(roles) != 0 {
		for i, f := range roles {
			roles[i] = strings.ToUpper(f)
		}
		joinRolesOnSQL = `role.name = ANY(` + args.Append(roles) + `)`
	}

	sql := `
		INSERT INTO role_permission(permission_id, role) (
			SELECT
				permission.id,
				role.name
			FROM permission
			JOIN role ON ` + joinRolesOnSQL + `
			WHERE permission.access_level = ` + args.Append(o.AccessLevel) + `
				AND permission.type = ` + args.Append(o.NodeType) + `
	` + andFieldsSQL + `
		) ON CONFLICT ON CONSTRAINT role_permission_pkey DO NOTHING
	`

	psName := preparedName("connectRoles", sql)

	_, err := prepareExec(db, psName, sql, args...)
	if err != nil {
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"operation": o,
		"fields":    fields,
		"roles":     roles,
	}).Info("granted permissions to roles")
	return nil
}

const deletePermissionSuiteSQL = `
	DELETE FROM permission
	WHERE type = $1
`

func DeletePermissionSuite(
	db Queryer,
	model interface{},
) error {
	mType, err := mytype.ParseNodeType(structs.Name(model))
	if err != nil {
		return err
	}
	mylog.Log.WithField(
		"node_type", mType,
	).Info("Perm.DeletePermissionSuite(node_type)")
	_, err = prepareExec(
		db,
		"deletePermissionSuite",
		deletePermissionSuiteSQL,
		mType,
	)
	if err != nil {
	}

	return nil
}

const getPermissionByRoleSQL = `
	SELECT
		access_level,
		created_at,
		field,
		granted_at,
		id,
		type,
		updated_at
	FROM
		role_permission_master
	WHERE role = $1
`

func GetByRole(
	db Queryer,
	role string,
) ([]Permission, error) {
	mylog.Log.Info("GetByRole()")
	permissions := make([]Permission, 0)

	rows, err := db.Query(getPermissionByRoleSQL, role)
	if err != nil {
		mylog.Log.WithError(err).Error("error during query")
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		p := new(Permission)
		err := rows.Scan(
			&p.AccessLevel,
			&p.CreatedAt,
			&p.GrantedAt,
			&p.ID,
			&p.Field,
			&p.Type,
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

var getQueryPermissionSQL = `
	SELECT
		access_level || ' ' || type AS operation,
		array_agg(field) AS fields
	FROM
		permission
	WHERE access_level = $1
		AND type = $2
		AND (audience = 'EVERYONE'
			OR id IN (
				SELECT permission_id
				FROM role_permission
				WHERE role = ANY($3)
			)
		)
	GROUP BY operation
`

func GetQueryPermission(
	db Queryer,
	o *mytype.Operation,
	roles []string,
) (*QueryPermission, error) {
	if o == nil {
		err := errors.New("operation is nil")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	p := &QueryPermission{}

	err := db.QueryRow(
		getQueryPermissionSQL,
		o.AccessLevel,
		o.NodeType,
		roles,
	).Scan(
		&p.Operation,
		&p.Fields,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	fields := make([]string, len(p.Fields.Elements))
	for i, f := range p.Fields.Elements {
		fields[i] = f.String
	}
	mylog.Log.WithFields(logrus.Fields{
		"fields":    fields,
		"operation": o,
		"roles":     roles,
	}).Info(util.Trace("query granted permission"))
	return p, nil
}

const updatePermissionSQL = `
	UPDATE permission
	SET audience = $1
	WHERE id = $2
`

func UpdatePermission(
	db Queryer,
	id string,
	a mytype.Audience,
) error {
	_, err := db.Exec(updatePermissionSQL, a, id)
	if err != nil {
		mylog.Log.WithError(err).Error("error during execution")
		return err
	}

	return nil
}

const updatePermissionAudienceSQL = `
	UPDATE permission
	SET audience = $1
	WHERE access_level = $2
		AND audience != $1
		AND type = $3
		AND field = ANY($4)
`

func UpdatePermissionAudience(
	db Queryer,
	o *mytype.Operation,
	a mytype.Audience,
	fields []string,
) error {
	mylog.Log.Info("UpdatePermissionAudience()")
	args := pgx.QueryArgs(make([]interface{}, 0, 4))

	andFieldsSQL := ""
	if len(fields) != 0 {
		for i, f := range fields {
			fields[i] = strings.ToLower(f)
		}
		andFieldsSQL = `AND field = ANY(` + args.Append(fields) + `)`
	}

	sql := `
		UPDATE permission
		SET audience = ` + args.Append(a) + `
		WHERE access_level = ` + args.Append(o.AccessLevel) + `
			AND audience != ` + args.Append(a) + `
			AND type = ` + args.Append(o.NodeType) + `
	` + andFieldsSQL

	psName := preparedName("updateOperationAudience", sql)

	_, err := prepareExec(db, psName, sql, args...)
	if err != nil {
		return err
	}

	mylog.Log.WithFields(logrus.Fields{
		"operation": o.String(),
		"audience":  a.String(),
		"fields":    fields,
	}).Info("updated operation audience")
	return nil
}

func getPermission(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) (*Permission, error) {
	var row Permission
	err := prepareQueryRow(db, name, sql, args...).Scan(
		&row.AccessLevel,
		&row.Audience,
		&row.CreatedAt,
		&row.Field,
		&row.ID,
		&row.Type,
		&row.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		mylog.Log.WithError(err).Error("failed to get permission")
		return nil, err
	}

	return &row, nil
}

func getManyPermission(
	db Queryer,
	name string,
	sql string,
	rows *[]*Permission,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Permission
		dbRows.Scan(
			&row.AccessLevel,
			&row.Audience,
			&row.CreatedAt,
			&row.Field,
			&row.ID,
			&row.Type,
			&row.UpdatedAt,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get permissions")
		return err
	}

	return nil
}
