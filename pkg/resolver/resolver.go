package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

var clientURL = "http://localhost:3000"

var InternalServerError = errors.New("something went wrong")

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

func eventPermitToResolver(ctx context.Context, event *repo.EventPermit, repos *repo.Repos) (interface{}, error) {
	eventType, err := event.Type()
	if err != nil {
		return nil, err
	}
	switch eventType {
	case data.CourseEvent:
		return courseEventPermitToResolver(event, repos)
	case data.LessonEvent:
		return lessonEventPermitToResolver(ctx, event, repos)
	case data.UserAssetEvent:
		return userAssetEventPermitToResolver(ctx, event, repos)
	case data.StudyEvent:
		return studyEventPermitToResolver(event, repos)
	default:
		return &eventResolver{Event: event, Repos: repos}, nil
	}
	return nil, nil
}

func courseEventPermitToResolver(event *repo.EventPermit, repos *repo.Repos) (interface{}, error) {
	payload := &data.CourseEventPayload{}
	eventPayload, err := event.Payload()
	if err != nil {
		return nil, err
	}
	if err := eventPayload.AssignTo(payload); err != nil {
		return nil, err
	}
	switch payload.Action {
	case data.CourseAppled:
		return &appledEventResolver{AppleableID: &payload.CourseID, Event: event, Repos: repos}, nil
	case data.CourseUnappled:
		return &unappledEventResolver{AppleableID: &payload.CourseID, Event: event, Repos: repos}, nil
	default:
		return &eventResolver{Event: event, Repos: repos}, nil
	}
}

func lessonEventPermitToResolver(ctx context.Context, event *repo.EventPermit, repos *repo.Repos) (interface{}, error) {
	payload := &data.LessonEventPayload{}
	eventPayload, err := event.Payload()
	if err != nil {
		return nil, err
	}
	if err := eventPayload.AssignTo(payload); err != nil {
		return nil, err
	}
	switch payload.Action {
	case data.LessonAddedToCourse:
		return nil, nil
	case data.LessonCreated:
		return &createdEventResolver{CreateableID: &payload.LessonID, Event: event, Repos: repos}, nil
	case data.LessonCommented:
		lessonComment, err := repos.LessonComment().Get(ctx, payload.CommentID.String)
		if err != nil {
			return nil, err
		}
		return &lessonCommentResolver{LessonComment: lessonComment, Repos: repos}, nil
	case data.LessonLabeled:
		return nil, nil
	case data.LessonMentioned:
		return nil, nil
	case data.LessonReferenced:
		return &referencedEventResolver{
			Event:           event,
			ReferenceableID: &payload.LessonID,
			Repos:           repos,
			SourceID:        &payload.SourceID,
		}, nil
	case data.LessonRemovedFromCourse:
		return nil, nil
	case data.LessonRenamed:
		return nil, nil
	case data.LessonUnlabeled:
		return nil, nil
	default:
		return &eventResolver{Event: event, Repos: repos}, nil
	}
}

func studyEventPermitToResolver(event *repo.EventPermit, repos *repo.Repos) (interface{}, error) {
	payload := &data.StudyEventPayload{}
	eventPayload, err := event.Payload()
	if err != nil {
		return nil, err
	}
	if err := eventPayload.AssignTo(payload); err != nil {
		return nil, err
	}
	switch payload.Action {
	case data.StudyCreated:
		return &createdEventResolver{CreateableID: &payload.StudyID, Event: event, Repos: repos}, nil
	case data.StudyAppled:
		return &appledEventResolver{AppleableID: &payload.StudyID, Event: event, Repos: repos}, nil
	case data.StudyUnappled:
		return &unappledEventResolver{AppleableID: &payload.StudyID, Event: event, Repos: repos}, nil
	default:
		return &eventResolver{Event: event, Repos: repos}, nil
	}
}

func userAssetEventPermitToResolver(ctx context.Context, event *repo.EventPermit, repos *repo.Repos) (interface{}, error) {
	payload := &data.UserAssetEventPayload{}
	eventPayload, err := event.Payload()
	if err != nil {
		return nil, err
	}
	if err := eventPayload.AssignTo(payload); err != nil {
		return nil, err
	}
	switch payload.Action {
	case data.UserAssetCreated:
		return &createdEventResolver{CreateableID: &payload.AssetID, Event: event, Repos: repos}, nil
	case data.UserAssetMentioned:
		return nil, nil
	case data.UserAssetReferenced:
		return &referencedEventResolver{
			Event:           event,
			ReferenceableID: &payload.AssetID,
			Repos:           repos,
			SourceID:        &payload.SourceID,
		}, nil
	case data.UserAssetRenamed:
		return nil, nil
	default:
		return &eventResolver{Event: event, Repos: repos}, nil
	}
}
