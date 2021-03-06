package resolver

import (
	"context"
	"errors"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type notificationResolver struct {
	Conf         *myconf.Config
	Notification *repo.NotificationPermit
	Repos        *repo.Repos
}

func (r *notificationResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Notification.CreatedAt()
	return graphql.Time{t}, err
}

func (r *notificationResolver) ID() (graphql.ID, error) {
	id, err := r.Notification.ID()
	return graphql.ID(id.String), err
}

func (r *notificationResolver) LastReadAt() (graphql.Time, error) {
	t, err := r.Notification.LastReadAt()
	return graphql.Time{t}, err
}

func (r *notificationResolver) Reason() (string, error) {
	return r.Notification.Reason()
}

func (r *notificationResolver) Subject(ctx context.Context) (*notificationSubjectResolver, error) {
	subjectID, err := r.Notification.SubjectID()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetNotificationSubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	subject, ok := resolver.(notificationSubject)
	if !ok {
		return nil, errors.New("cannot convert resolver to notification subject")
	}
	return &notificationSubjectResolver{subject}, nil
}

func (r *notificationResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Notification.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *notificationResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Notification.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *notificationResolver) User(ctx context.Context) (*userResolver, error) {
	userID, err := r.Notification.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
