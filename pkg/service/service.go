package service

import "github.com/marksauter/markus-ninja-api/pkg/mydb"

type Services struct {
	Auth *AuthService
	Perm *PermService
	Role *RoleService
	User *UserService
}

func NewServices(db *mydb.DB) *Services {
	return &Services{
		Auth: NewAuthService(),
		Perm: NewPermService(db),
		Role: NewRoleService(db),
		User: NewUserService(db),
	}
}
