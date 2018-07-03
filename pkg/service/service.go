package service

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myaws"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
)

type Services struct {
	Appled        *data.AppledService
	Auth          *AuthService
	Email         *data.EmailService
	Enrolled      *data.EnrolledService
	Event         *data.EventService
	EVT           *data.EVTService
	Label         *data.LabelService
	Labeled       *data.LabeledService
	Lesson        *data.LessonService
	LessonComment *data.LessonCommentService
	Mail          *MailService
	Notification  *data.NotificationService
	Perm          *data.PermissionService
	PRT           *data.PRTService
	Role          *data.RoleService
	Storage       *StorageService
	Study         *data.StudyService
	Topic         *data.TopicService
	Topiced       *data.TopicedService
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
		Appled:        data.NewAppledService(db),
		Auth:          NewAuthService(myaws.NewKMS(), authConfig),
		Email:         data.NewEmailService(db),
		Enrolled:      data.NewEnrolledService(db),
		Event:         data.NewEventService(db),
		EVT:           data.NewEVTService(db),
		Label:         data.NewLabelService(db),
		Labeled:       data.NewLabeledService(db),
		Lesson:        data.NewLessonService(db),
		LessonComment: data.NewLessonCommentService(db),
		Mail:          NewMailService(myaws.NewSES(), mailConfig),
		Notification:  data.NewNotificationService(db),
		Perm:          data.NewPermissionService(db),
		PRT:           data.NewPRTService(db),
		Role:          data.NewRoleService(db),
		Storage:       storageSvc,
		Study:         data.NewStudyService(db),
		Topic:         data.NewTopicService(db),
		Topiced:       data.NewTopicedService(db),
		User:          data.NewUserService(db),
		UserAsset:     data.NewUserAssetService(db),
	}, nil
}
