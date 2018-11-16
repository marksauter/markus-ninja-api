package data

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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
	GrantedAt   pgtype.Timestamptz `db:"granted_at"`
	Name        pgtype.Varchar     `db:"name"`
	UserID      mytype.OID         `db:"user_id"`
}

func getManyRole(
	db Queryer,
	name string,
	sql string,
	rows *[]*Role,
	args ...interface{},
) error {
	dbRows, err := prepareQuery(db, name, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Role
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.Name,
		)
		*rows = append(*rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Debug(util.Trace(""))
		return err
	}

	return nil
}

func GetRoleByUser(
	db Queryer,
	userID string,
	po *PageOptions,
) ([]*Role, error) {
	var rows []*Role
	if po != nil && po.Limit() > 0 {
		limit := po.Limit()
		if limit > 0 {
			rows = make([]*Role, 0, limit)
		} else {
			mylog.Log.Info(util.Trace("limit is 0"))
			return rows, nil
		}
	}

	var args pgx.QueryArgs
	where := func(string) string { return "" }

	selects := []string{
		"created_at",
		"description",
		"granted_at",
		"name",
		"user_id",
	}
	from := "user_role_master"
	sql := SQL3(selects, from, where, nil, &args, po)

	psName := preparedName("getRoleByUser", sql)

	dbRows, err := prepareQuery(db, psName, sql, args...)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var row Role
		dbRows.Scan(
			&row.CreatedAt,
			&row.Description,
			&row.GrantedAt,
			&row.Name,
			&row.UserID,
		)
		rows = append(rows, &row)
	}

	if err := dbRows.Err(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	mylog.Log.WithField("n", len(rows)).Info(util.Trace("roles found"))
	return rows, nil
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
	userID string,
	roles ...string,
) error {
	if len(roles) > 0 {
		_, err := prepareExec(
			db,
			"grantUserRoles",
			grantUserRolesSQL,
			roles,
			userID,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return err
		}
	}

	mylog.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"roles":   roles,
	}).Info(util.Trace("roles granted"))
	return nil
}
