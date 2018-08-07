package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

var clientURL = "http://localhost:3000"

type RootResolver struct {
	Repos *repo.Repos
	Svcs  *service.Services
}

func nodePermitToResolver(p repo.NodePermit, repos *repo.Repos) (interface{}, error) {
	id, err := p.ID()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Course":
		course, ok := p.(*repo.CoursePermit)
		if !ok {
			return nil, errors.New("cannot convert permit to course")
		}
		return &courseResolver{Course: course, Repos: repos}, nil
	case "Email":
		email, ok := p.(*repo.EmailPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to email")
		}
		return &emailResolver{Email: email, Repos: repos}, nil
	case "Event":
		event, ok := p.(*repo.EventPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to event")
		}
		return &eventResolver{Event: event, Repos: repos}, nil
	case "Label":
		label, ok := p.(*repo.LabelPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to label")
		}
		return &labelResolver{Label: label, Repos: repos}, nil
	case "Lesson":
		lesson, ok := p.(*repo.LessonPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to lesson")
		}
		return &lessonResolver{Lesson: lesson, Repos: repos}, nil
	case "LessonComment":
		lessonComment, ok := p.(*repo.LessonCommentPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to lessonComment")
		}
		return &lessonCommentResolver{LessonComment: lessonComment, Repos: repos}, nil
	case "Notification":
		notification, ok := p.(*repo.NotificationPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to notification")
		}
		return &notificationResolver{Notification: notification, Repos: repos}, nil
	case "Study":
		study, ok := p.(*repo.StudyPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to study")
		}
		return &studyResolver{Study: study, Repos: repos}, nil
	case "Topic":
		topic, ok := p.(*repo.TopicPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to topic")
		}
		return &topicResolver{Topic: topic, Repos: repos}, nil
	case "User":
		user, ok := p.(*repo.UserPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to user")
		}
		return &userResolver{User: user, Repos: repos}, nil
	case "UserAsset":
		userAsset, ok := p.(*repo.UserAssetPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to userAsset")
		}
		return &userAssetResolver{UserAsset: userAsset, Repos: repos}, nil
	}
	return nil, nil
}

func eventPermitToResolver(event *repo.EventPermit, repos *repo.Repos) (interface{}, error) {
	action, err := event.Action()
	if err != nil {
		return nil, err
	}
	switch action {
	case data.AppledEvent:
		return &appledEventResolver{Event: event, Repos: repos}, nil
	case data.CreatedEvent:
		return &createdEventResolver{Event: event, Repos: repos}, nil
	case data.CommentedEvent:
		return &commentedEventResolver{Event: event, Repos: repos}, nil
	case data.ReferencedEvent:
		return &referencedEventResolver{Event: event, Repos: repos}, nil
	default:
		return &eventResolver{Event: event, Repos: repos}, nil
	}
	return nil, nil
}
