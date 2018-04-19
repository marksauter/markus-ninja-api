package service

import (
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
		err := rows.Scan(&r.ID, &r.Name, &r.CreatedAt)
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
