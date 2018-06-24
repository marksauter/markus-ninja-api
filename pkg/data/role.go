package data

import (
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type Role struct {
	CreatedAt   pgtype.Timestamptz `db:"created_at"`
	Description pgtype.Text        `db:"description"`
	Name        mytype.RoleName    `db:"name"`
}

type RoleService struct {
	db Queryer
}

func NewRoleService(db Queryer) *RoleService {
	return &RoleService{db}
}

func (s *RoleService) getMany(name string, sql string, args ...interface{}) ([]*Role, error) {
	var rows []*Role

	dbRows, err := prepareQuery(s.db, name, sql, args...)
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

	mylog.Log.WithField("n", len(rows)).Info("")

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

func (s *RoleService) GetByUser(userId string) ([]*Role, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("Role.GetByUser(user_id)")
	return s.getMany("getRolesByUser", getRolesByUserSQL, userId)
}

const grantUserRolesSQL = `
	INSERT INTO user_role (user_id, role)
	SELECT DISTINCT a.id, r.name
	FROM account a
	INNER JOIN role r ON r.name = ANY($1)
	WHERE a.id = $2
	ON CONFLICT ON CONSTRAINT user_role_pkey DO NOTHING
`

func (s *RoleService) GrantUser(userId string, roles ...mytype.RoleNameValue) error {
	mylog.Log.WithFields(logrus.Fields{
		"user_id": userId,
		"roles":   roles,
	}).Info("GrantUser(user_id, roles)")
	if len(roles) > 0 {
		roleArgs := make([]string, len(roles))
		for i, r := range roles {
			roleArgs[i] = r.String()
		}
		_, err := prepareExec(
			s.db,
			"grantUserRoles",
			grantUserRolesSQL,
			roleArgs,
			userId,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error during execution")
			return err
		}
	}

	return nil
}
