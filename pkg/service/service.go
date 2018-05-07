package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
)

type Services struct {
	Auth *AuthService
	AVT  *data.EmailVerificationTokenService
	Mail *MailService
	Perm *data.PermService
	PWRT *data.PasswordResetTokenService
	Role *data.RoleService
	User *data.UserService
}

func NewServices(conf *myconf.Config, db data.Queryer) *Services {
	mailConfig := &MailServiceConfig{
		Host:        conf.SMTPHost,
		Port:        conf.SMTPPort,
		User:        conf.SMTPUser,
		Password:    conf.SMTPPassword,
		FromAddress: conf.SMTPFromAddr,
		RootURL:     conf.SMTPRootURL,
	}
	return &Services{
		Auth: NewAuthService(),
		AVT:  data.NewEmailVerificationTokenService(db),
		Mail: NewMailService(mailConfig),
		Perm: data.NewPermService(db),
		PWRT: data.NewPasswordResetTokenService(db),
		Role: data.NewRoleService(db),
		User: data.NewUserService(db),
	}
}
