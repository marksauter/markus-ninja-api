package service

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type RoleService struct {
	db     *sqlx.DB
	logger *mylog.Logger
}

func NewRoleService(db *sqlx.DB, logger *mylog.Logger) *RoleService {
	return &RoleService{db: db, logger: logger}
}

func (s *RoleService) GetByUserId(userId string) ([]*model.Role, error) {
	roles := make([]*model.Role, 0)

	roleSQL := `SELECT role.*
	FROM roles role
	INNER JOIN rel_users_roles ur ON role.id = ur.role_id
	WHERE ur.user_id = $1`
	err := s.db.Select(&roles, roleSQL, userId)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return roles, nil
		default:
			return nil, err
		}
	}
	return roles, nil
}
