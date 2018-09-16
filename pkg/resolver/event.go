package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Event = eventResolver

type eventResolver struct {
	Event *repo.EventPermit
	Repos *repo.Repos
}

func (r *eventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *eventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *eventResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Event.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *eventResolver) User(ctx context.Context) (*userResolver, error) {
	userID, err := r.Event.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}
