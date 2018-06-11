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
	Storage       *StorageService
	Study         *data.StudyService
	Topic         *data.TopicService
	User          *data.UserService
	UserAsset     *data.UserAssetService
}

func NewServices(conf *myconf.Config, db data.Queryer) (*Services, error) {
	authConfig := &AuthServiceConfig{
		KeyId: conf.AuthKeyId,
	}
	mailConfig := &MailServiceConfig{
		CharSet: conf.MailCharSet,
		Sender:  conf.MailSender,
		RootURL: conf.MailRootURL,
	}
	storageSvc, err := NewStorageService()
	if err != nil {
		return nil, err
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
		Storage:       storageSvc,
		Study:         data.NewStudyService(db),
		Topic:         data.NewTopicService(db),
		User:          data.NewUserService(db),
		UserAsset:     data.NewUserAssetService(db),
	}, nil
}
