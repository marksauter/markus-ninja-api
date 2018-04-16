package service

import (
	"database/sql"

	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
)

type RoleService struct {
	db *mydb.DB
}

func NewRoleService(db *mydb.DB) *RoleService {
	return &RoleService{db: db}
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
