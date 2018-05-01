package data

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

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

func (s *RoleService) Create(name string) (*RoleModel, error) {
	mylog.Log.WithField("name", name).Info("Create(name) Role")
	roleId := oid.New("Role")
	roleSQL := `
		INSERT INTO role(id, name)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT role_name_key DO NOTHING
		RETURNING
			id,
			name,
			created_at,
			updated_at
	`
	row := s.db.QueryRow(roleSQL, roleId, name)
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
		INNER JOIN account_role ar ON role.id = ar.role_id
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
