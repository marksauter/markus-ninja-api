package service

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/marksauter/markus-ninja-api/pkg/attr"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type RoleService struct {
	db *mydb.DB
}

func NewRoleService(db *mydb.DB) *RoleService {
	return &RoleService{db: db}
}

func (s *RoleService) Create(name string) (*model.Role, error) {
	mylog.Log.WithField("name", name).Info("Create(name) Role")
	roleId := attr.NewId("Role")
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
	row := s.db.QueryRow(roleSQL, roleId.String(), name)
	role := new(model.Role)
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
			switch mydb.PSQLError(pgErr.Code) {
			default:
				return nil, err
			case mydb.UniqueViolation:
				return nil, fmt.Errorf(`role "%v" already exists`, name)
			}
		}
	}

	mylog.Log.Debug("role created")
	return role, nil
}

func (s *RoleService) GetByUserId(userId string) ([]model.Role, error) {
	roles := make([]model.Role, 0)

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
