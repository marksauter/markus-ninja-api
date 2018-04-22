package service

import "github.com/marksauter/markus-ninja-api/pkg/mydb"

type Services struct {
	Auth *AuthService
	Perm *PermissionService
	Role *RoleService
	User *UserService
}

func NewServices(db *mydb.DB) *Services {
	roleSvc := NewRoleService(db)
	return &Services{
		Auth: NewAuthService(),
		Perm: NewPermissionService(db),
		Role: roleSvc,
		User: NewUserService(db, roleSvc),
	}
}
