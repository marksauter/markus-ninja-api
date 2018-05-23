package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
)

type Services struct {
	Auth          *AuthService
	Email         *data.EmailService
	EVT           *data.EVTService
	Lesson        *data.LessonService
	LessonComment *data.LessonCommentService
	Mail          *MailService
	Perm          *data.PermService
	PWRT          *data.PasswordResetTokenService
	Role          *data.RoleService
	Study         *data.StudyService
	User          *data.UserService
	UserEmail     *data.UserEmailService
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
		Auth:          NewAuthService(),
		EVT:           data.NewEVTService(db),
		Email:         data.NewEmailService(db),
		Lesson:        data.NewLessonService(db),
		LessonComment: data.NewLessonCommentService(db),
		Mail:          NewMailService(mailConfig),
		Perm:          data.NewPermService(db),
		PWRT:          data.NewPasswordResetTokenService(db),
		Role:          data.NewRoleService(db),
		Study:         data.NewStudyService(db),
		User:          data.NewUserService(db),
		UserEmail:     data.NewUserEmailService(db),
	}
}
