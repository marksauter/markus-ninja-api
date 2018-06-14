package data

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type RoleType int

const (
	AdminRole RoleType = iota
	MemberRole
	OwnerRole
	UserRole
)

func (r RoleType) String() string {
	switch r {
	case AdminRole:
		return "ADMIN"
	case MemberRole:
		return "MEMBER"
	case OwnerRole:
		return "OWNER"
	case UserRole:
		return "USER"
	default:
		return "unknown"
	}
}

type Role struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RoleService struct {
	db Queryer
}

func NewRoleService(db Queryer) *RoleService {
	return &RoleService{db}
}

var createRoleSQL = `
	INSERT INTO role(id, name)
	VALUES ($1, $2)
	ON CONFLICT ON CONSTRAINT role_name_key DO NOTHING
	RETURNING
		created_at,
		id,
		name,
		updated_at
`

func (s *RoleService) Create(name string) (*Role, error) {
	mylog.Log.WithField("name", name).Info("Role.Create(name)")
	id, _ := mytype.NewOID("Role")
	row := s.db.QueryRow(createRoleSQL, id, name)
	role := new(Role)
	err := row.Scan(
		&role.CreatedAt,
		&role.Id,
		&role.Name,
		&role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return role, nil
		}
		mylog.Log.WithField("error", err).Error("error during scan")
		if pgErr, ok := err.(pgx.PgError); ok {
			switch PSQLError(pgErr.Code) {
			default:
				return nil, err
			case UniqueViolation:
				return nil, fmt.Errorf(`role "%v" already exists`, name)
			}
		}
	}

	mylog.Log.Debug("role created")
	return role, nil
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
			&row.Id,
			&row.Name,
			&row.UpdatedAt,
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
		id,
		name
	FROM
		role r
	INNER JOIN user_role ur ON ur.role_id = r.id
	WHERE ur.user_id = $1
`

func (s *RoleService) GetByUser(userId string) ([]*Role, error) {
	mylog.Log.WithField(
		"user_id", userId,
	).Info("Role.GetByUser(user_id)")
	return s.getMany("getRolesByUser", getRolesByUserSQL, userId)
}

const grantUserRolesSQL = `
	INSERT INTO user_role (user_id, role_id)
	SELECT DISTINCT a.id, r.id
	FROM account a
	INNER JOIN role r ON r.name = ANY($1)
	WHERE a.id = $2
	ON CONFLICT ON CONSTRAINT user_role_pkey DO NOTHING
`

func (s *RoleService) GrantUser(userId string, roles ...RoleType) error {
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
