package data

import "github.com/marksauter/markus-ninja-api/pkg/mydb"

type Services struct {
	Auth *AuthService
	AVT  *AccountVerificationTokenService
	Perm *PermService
	PWRT *PasswordResetTokenService
	Role *RoleService
	User *UserService
}

func NewServices(db *mydb.DB) *Services {
	return &Services{
		Auth: NewAuthService(),
		AVT:  NewAccountVerificationTokenService(db),
		Perm: NewPermService(db),
		PWRT: NewPasswordResetTokenService(db),
		Role: NewRoleService(db),
		User: NewUserService(db),
	}
}
