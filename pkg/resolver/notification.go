package resolver

import (
	"context"

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

func (r *notificationResolver) Subject(ctx context.Context) (*lessonResolver, error) {
	subjectId, err := r.Notification.SubjectId()
	if err != nil {
		return nil, err
	}
	subject, err := r.Repos.Lesson().Get(ctx, subjectId.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: subject, Repos: r.Repos}, nil
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

func (r *notificationResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Notification.UpdatedAt()
	return graphql.Time{t}, err
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
