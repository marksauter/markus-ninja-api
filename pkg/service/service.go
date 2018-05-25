package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
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
	PRT           *data.PRTService
	Role          *data.RoleService
	Study         *data.StudyService
	User          *data.UserService
}

func NewServices(conf *myconf.Config, db data.Queryer) *Services {
	authConfig := &AuthServiceConfig{
		KeyId: conf.AuthKeyId,
	}
	mailConfig := &MailServiceConfig{
		CharSet: conf.MailCharSet,
		Sender:  conf.MailSender,
		RootURL: conf.MailRootURL,
	}
	return &Services{
		Auth:          NewAuthService(myaws.NewKMS(), authConfig),
		EVT:           data.NewEVTService(db),
		Email:         data.NewEmailService(db),
		Lesson:        data.NewLessonService(db),
		LessonComment: data.NewLessonCommentService(db),
		Mail:          NewMailService(myaws.NewSES(), mailConfig),
		Perm:          data.NewPermService(db),
		PRT:           data.NewPRTService(db),
		Role:          data.NewRoleService(db),
		Study:         data.NewStudyService(db),
		User:          data.NewUserService(db),
	}
}
