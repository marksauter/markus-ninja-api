package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

var InternalServerError = errors.New("something went wrong")

type RootResolver struct {
	Conf  *myconf.Config
	Repos *repo.Repos
	Svcs  *service.Services
}

func nodePermitToResolver(
	p repo.NodePermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
	id, err := p.ID()
	if err != nil {
		return nil, err
	}
	switch id.Type {
	case "Activity":
		activity, ok := p.(*repo.ActivityPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to activity")
		}
		return &activityResolver{Activity: activity, Conf: conf, Repos: repos}, nil
	case "Comment":
		comment, ok := p.(*repo.CommentPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to comment")
		}
		return &commentResolver{Comment: comment, Conf: conf, Repos: repos}, nil
	case "Course":
		course, ok := p.(*repo.CoursePermit)
		if !ok {
			return nil, errors.New("cannot convert permit to course")
		}
		return &courseResolver{Course: course, Conf: conf, Repos: repos}, nil
	case "Email":
		email, ok := p.(*repo.EmailPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to email")
		}
		return &emailResolver{Email: email, Conf: conf, Repos: repos}, nil
	case "Event":
		event, ok := p.(*repo.EventPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to event")
		}
		return &eventResolver{Event: event, Conf: conf, Repos: repos}, nil
	case "Label":
		label, ok := p.(*repo.LabelPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to label")
		}
		return &labelResolver{Label: label, Conf: conf, Repos: repos}, nil
	case "Lesson":
		lesson, ok := p.(*repo.LessonPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to lesson")
		}
		return &lessonResolver{Lesson: lesson, Conf: conf, Repos: repos}, nil
	case "Notification":
		notification, ok := p.(*repo.NotificationPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to notification")
		}
		return &notificationResolver{Notification: notification, Conf: conf, Repos: repos}, nil
	case "Study":
		study, ok := p.(*repo.StudyPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to study")
		}
		return &studyResolver{Study: study, Conf: conf, Repos: repos}, nil
	case "Topic":
		topic, ok := p.(*repo.TopicPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to topic")
		}
		return &topicResolver{Topic: topic, Conf: conf, Repos: repos}, nil
	case "User":
		user, ok := p.(*repo.UserPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to user")
		}
		return &userResolver{User: user, Conf: conf, Repos: repos}, nil
	case "UserAsset":
		userAsset, ok := p.(*repo.UserAssetPermit)
		if !ok {
			return nil, errors.New("cannot convert permit to userAsset")
		}
		return &userAssetResolver{UserAsset: userAsset, Conf: conf, Repos: repos}, nil
	}
	return nil, nil
}

func eventPermitToResolver(
	ctx context.Context,
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
	eventType, err := event.Type()
	if err != nil {
		return nil, err
	}
	switch eventType {
	case data.ActivityEvent:
		return activityEventPermitToResolver(event, repos, conf)
	case data.CourseEvent:
		return courseEventPermitToResolver(event, repos, conf)
	case data.LessonEvent:
		return lessonEventPermitToResolver(ctx, event, repos, conf)
	case data.UserAssetEvent:
		return userAssetEventPermitToResolver(ctx, event, repos, conf)
	case data.StudyEvent:
		return studyEventPermitToResolver(event, repos, conf)
	default:
		return nil, nil
	}
	return nil, nil
}

func activityEventPermitToResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
	payload := &data.ActivityEventPayload{}
	eventPayload, err := event.Payload()
	if err != nil {
		return nil, err
	}
	if err := eventPayload.AssignTo(payload); err != nil {
		return nil, err
	}
	switch payload.Action {
	case data.ActivityCreated:
		return &createdEventResolver{
			CreateableID: &payload.ActivityID,
			Conf:         conf,
			Event:        event,
			Repos:        repos,
		}, nil
	default:
		return nil, nil
	}
}

func courseEventPermitToResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
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
		return &appledEventResolver{
			AppleableID: &payload.CourseID,
			Conf:        conf,
			Event:       event,
			Repos:       repos,
		}, nil
	case data.CourseCreated:
		return &createdEventResolver{
			CreateableID: &payload.CourseID,
			Conf:         conf,
			Event:        event,
			Repos:        repos,
		}, nil
	case data.CoursePublished:
		return &publishedEventResolver{
			Conf:          conf,
			Event:         event,
			PublishableID: &payload.CourseID,
			Repos:         repos,
		}, nil
	case data.CourseUnappled:
		return &unappledEventResolver{
			AppleableID: &payload.CourseID,
			Conf:        conf,
			Event:       event,
			Repos:       repos,
		}, nil
	default:
		return nil, nil
	}
}

func lessonEventPermitToResolver(
	ctx context.Context,
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
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
		return &addedToCourseEventResolver{
			Conf:     conf,
			CourseID: &payload.CourseID,
			Event:    event,
			LessonID: &payload.LessonID,
			Repos:    repos,
		}, nil
	case data.LessonCreated:
		return &createdEventResolver{
			Conf:         conf,
			CreateableID: &payload.LessonID,
			Event:        event,
			Repos:        repos,
		}, nil
	case data.LessonCommented:
		comment, err := repos.Comment().Get(
			ctx,
			payload.CommentID.String,
		)
		if err != nil {
			return nil, err
		}
		return &commentResolver{
			Conf:    conf,
			Comment: comment,
			Repos:   repos,
		}, nil
	case data.LessonLabeled:
		return &labeledEventResolver{
			Conf:        conf,
			LabelID:     &payload.LabelID,
			LabelableID: &payload.LessonID,
			Event:       event,
			Repos:       repos,
		}, nil
	case data.LessonMentioned:
		return nil, nil
	case data.LessonPublished:
		return &publishedEventResolver{
			Conf:          conf,
			Event:         event,
			PublishableID: &payload.LessonID,
			Repos:         repos,
		}, nil
	case data.LessonReferenced:
		return &referencedEventResolver{
			Conf:            conf,
			Event:           event,
			ReferenceableID: &payload.LessonID,
			Repos:           repos,
			SourceID:        &payload.SourceID,
		}, nil
	case data.LessonRemovedFromCourse:
		return &removedFromCourseEventResolver{
			Conf:     conf,
			CourseID: &payload.CourseID,
			Event:    event,
			LessonID: &payload.LessonID,
			Repos:    repos,
		}, nil
	case data.LessonRenamed:
		return &renamedEventResolver{
			Conf:         conf,
			RenameableID: &payload.LessonID,
			Rename:       &payload.Rename,
			Event:        event,
			Repos:        repos,
		}, nil
	case data.LessonUnlabeled:
		return &unlabeledEventResolver{
			Conf:        conf,
			LabelID:     &payload.LabelID,
			LabelableID: &payload.LessonID,
			Event:       event,
			Repos:       repos,
		}, nil
	default:
		return nil, nil
	}
}

func studyEventPermitToResolver(
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
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
		return &createdEventResolver{
			Conf:         conf,
			CreateableID: &payload.StudyID,
			Event:        event,
			Repos:        repos,
		}, nil
	case data.StudyAppled:
		return &appledEventResolver{
			AppleableID: &payload.StudyID,
			Conf:        conf,
			Event:       event,
			Repos:       repos,
		}, nil
	case data.StudyUnappled:
		return &unappledEventResolver{
			AppleableID: &payload.StudyID,
			Conf:        conf,
			Event:       event,
			Repos:       repos,
		}, nil
	default:
		return nil, nil
	}
}

func userAssetEventPermitToResolver(
	ctx context.Context,
	event *repo.EventPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (interface{}, error) {
	payload := &data.UserAssetEventPayload{}
	eventPayload, err := event.Payload()
	if err != nil {
		return nil, err
	}
	if err := eventPayload.AssignTo(payload); err != nil {
		return nil, err
	}
	switch payload.Action {
	case data.UserAssetAddedToActivity:
		return &addedToActivityEventResolver{
			Conf:       conf,
			ActivityID: &payload.ActivityID,
			Event:      event,
			AssetID:    &payload.AssetID,
			Repos:      repos,
		}, nil
	case data.UserAssetCommented:
		comment, err := repos.Comment().Get(
			ctx,
			payload.CommentID.String,
		)
		if err != nil {
			return nil, err
		}
		return &commentResolver{
			Conf:    conf,
			Comment: comment,
			Repos:   repos,
		}, nil
	case data.UserAssetCreated:
		return &createdEventResolver{
			Conf:         conf,
			CreateableID: &payload.AssetID,
			Event:        event,
			Repos:        repos,
		}, nil
	case data.UserAssetLabeled:
		return &labeledEventResolver{
			Conf:        conf,
			LabelID:     &payload.LabelID,
			LabelableID: &payload.AssetID,
			Event:       event,
			Repos:       repos,
		}, nil
	case data.UserAssetMentioned:
		return nil, nil
	case data.UserAssetReferenced:
		return &referencedEventResolver{
			Conf:            conf,
			Event:           event,
			ReferenceableID: &payload.AssetID,
			Repos:           repos,
			SourceID:        &payload.SourceID,
		}, nil
	case data.UserAssetRemovedFromActivity:
		return &removedFromActivityEventResolver{
			Conf:       conf,
			ActivityID: &payload.ActivityID,
			Event:      event,
			AssetID:    &payload.AssetID,
			Repos:      repos,
		}, nil
	case data.UserAssetRenamed:
		return &renamedEventResolver{
			Conf:         conf,
			Event:        event,
			RenameableID: &payload.AssetID,
			Rename:       &payload.Rename,
			Repos:        repos,
		}, nil
	case data.UserAssetUnlabeled:
		return &unlabeledEventResolver{
			Conf:        conf,
			LabelID:     &payload.LabelID,
			LabelableID: &payload.AssetID,
			Event:       event,
			Repos:       repos,
		}, nil
	default:
		return nil, nil
	}
}
