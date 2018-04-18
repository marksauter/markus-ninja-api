package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/mydb"
)

type RoleService struct {
	db *mydb.DB
}

func NewRoleService(db *mydb.DB) *RoleService {
	return &RoleService{db: db}
}

// func (s *RoleService) GetByUserId(userId string) ([]*model.Role, error) {
//   roles := make([]*model.Role, 0)
//
//   roleSQL := `
//     SELECT
//       *
//     FROM
//       role
//     INNER JOIN account_role ar ON role.id = ar.role_id
//     WHERE ar.user_id = $1
//   `
//   err := s.db.Select(&roles, roleSQL, userId)
//   if err != nil {
//     switch err {
//     case sql.ErrNoRows:
//       return roles, nil
//     default:
//       return nil, err
//     }
//   }
//   return roles, nil
// }
