package data

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/sirupsen/logrus"
)

type Role int

const (
	AdminRole Role = iota
	MemberRole
	OwnerRole
	UserRole
)

func (r Role) String() string {
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

type RoleModel struct {
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
		id,
		name,
		created_at,
		updated_at
`

func (s *RoleService) Create(name string) (*RoleModel, error) {
	mylog.Log.WithField("name", name).Info("Create(name) Role")
	id, _ := mytype.NewOID("Role")
	row := s.db.QueryRow(createRoleSQL, id, name)
	role := new(RoleModel)
	err := row.Scan(
		&role.Id,
		&role.Name,
		&role.CreatedAt,
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

func (s *RoleService) GetByUserId(userId string) ([]RoleModel, error) {
	roles := make([]RoleModel, 0)

	roleSQL := `
		SELECT
			id,
			name,
			created_at,
		FROM
			role
		INNER JOIN user_role ar ON role.id = ar.role_id
		WHERE ar.user_id = $1
	`
	rows, err := s.db.Query(roleSQL, userId)
	if err != nil {
		mylog.Log.WithField("error", err).Error("error during query")
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		r := roles[i]
		err := rows.Scan(&r.Id, &r.Name, &r.CreatedAt)
		if err != nil {
			mylog.Log.WithField("error", err).Error("error during scan")
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		mylog.Log.WithField("error", err).Error("error during rows processing")
		return nil, err
	}

	mylog.Log.Debug("roles found")
	return roles, nil
}

const grantUserRolesSQL = `
	INSERT INTO user_role (user_id, role_id)
	SELECT DISTINCT a.id, r.id
	FROM account a
	INNER JOIN role r ON r.name = ANY($1)
	WHERE a.id = $2
	ON CONFLICT ON CONSTRAINT user_role_pkey DO NOTHING
`

func (s *RoleService) GrantUser(userId string, roles ...Role) error {
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
