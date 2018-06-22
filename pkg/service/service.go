package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
)

type Services struct {
	Auth          *AuthService
	Email         *data.EmailService
	Event         *data.EventService
	EVT           *data.EVTService
	Lesson        *data.LessonService
	LessonComment *data.LessonCommentService
	Mail          *MailService
	Perm          *data.PermService
	PRT           *data.PRTService
	Role          *data.RoleService
	Storage       *StorageService
	Study         *data.StudyService
	StudyApple    *data.StudyAppleService
	StudyEnroll   *data.StudyEnrollService
	Topic         *data.TopicService
	User          *data.UserService
	UserAsset     *data.UserAssetService
	UserEnroll     *data.UserEnrollService
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
		Email:         data.NewEmailService(db),
		Event:         data.NewEventService(db),
		EVT:           data.NewEVTService(db),
		Lesson:        data.NewLessonService(db),
		LessonComment: data.NewLessonCommentService(db),
		Mail:          NewMailService(myaws.NewSES(), mailConfig),
		Perm:          data.NewPermService(db),
		PRT:           data.NewPRTService(db),
		Role:          data.NewRoleService(db),
		Storage:       storageSvc,
		Study:         data.NewStudyService(db),
		StudyApple:    data.NewStudyAppleService(db),
		StudyEnroll:   data.NewStudyEnrollService(db),
		Topic:         data.NewTopicService(db),
		User:          data.NewUserService(db),
		UserAsset:     data.NewUserAssetService(db),
		UserEnroll:     data.NewUserEnrollService(db),
	}, nil
}
