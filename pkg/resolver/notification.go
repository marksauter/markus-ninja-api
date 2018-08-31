package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Notification = notificationResolver

type notificationResolver struct {
	Notification *repo.NotificationPermit
	Repos        *repo.Repos
}

func (r *notificationResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Notification.CreatedAt()
	return graphql.Time{t}, err
}

func (r *notificationResolver) Event(ctx context.Context) (*notificationEventResolver, error) {
	eventId, err := r.Notification.EventId()
	if err != nil {
		return nil, err
	}
	e, err := r.Repos.Event().Get(ctx, eventId.String)
	if err != nil {
		return nil, err
	}
	resolver, err := eventPermitToResolver(ctx, e, r.Repos)
	if err != nil {
		return nil, err
	}
	event, ok := resolver.(notificationEvent)
	if !ok {
		return nil, errors.New("cannot convert resolver to notification event")
	}
	return &notificationEventResolver{event}, nil
}

func (r *notificationResolver) ID() (graphql.ID, error) {
	id, err := r.Notification.ID()
	return graphql.ID(id.String), err
}

func (r *notificationResolver) Reason() (string, error) {
	return r.Notification.Reason()
}

func (r *notificationResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.Notification.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *notificationResolver) User(ctx context.Context) (*userResolver, error) {
	userId, err := r.Notification.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}
