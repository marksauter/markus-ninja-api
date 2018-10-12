package data

import (
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/sirupsen/logrus"
)

const (
	AdminRole  = "ADMIN"
	MemberRole = "MEMBER"
	OwnerRole  = "OWNER"
	UserRole   = "USER"
)

type Role struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at"`
	Description pgtype.Text        `db:"description"`
	Name        pgtype.Varchar     `db:"name"`
}

func getManyRole(
	db Queryer,
	name string,
	sql string,
	args ...interface{},
) ([]*Role, error) {
	var rows []*Role

	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		return nil, err
	}

	for dbRows.Next() {
		var row Role
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.Name,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error("failed to get roles")
		return nil, err
	}

	return rows, nil
}

const getRolesByUserSQL = `
	SELECT
		created_at,
		description,
		name
	FROM
		role r
	INNER JOIN user_role ur ON ur.role = r.name
	WHERE ur.user_id = $1
`

func GetRoleByUser(
	db Queryer,
	userId string,
) ([]*Role, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("GetRoleByUser(user_id)")
	return getManyRole(db, "getRolesByUser", getRolesByUserSQL, userId)
}

const grantUserRolesSQL = `
	INSERT INTO user_role (user_id, role)
	SELECT DISTINCT a.id, r.name
	FROM account a
	INNER JOIN role r ON r.name = ANY($1)
	WHERE a.id = $2
	ON CONFLICT ON CONSTRAINT user_role_pkey DO NOTHING
`

func GrantUserRoles(
	db Queryer,
	userId string,
	roles ...string,
) error {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"roles":   roles,
	}).Info("GrantUserRoles(user_id, roles)")
	if len(roles) > 0 {
		_, err := prepareExec(
			db,
			"grantUserRoles",
			grantUserRolesSQL,
			roles,
			userId,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error during execution")
			return err
		}
	}

	return nil
}
