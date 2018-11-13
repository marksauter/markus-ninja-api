package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createdEventResolver struct {
	Conf         *myconf.Config
	CreateableID *mytype.OID
	Event        *repo.EventPermit
	Repos        *repo.Repos
}

func (r *createdEventResolver) Createable(ctx context.Context) (*createableResolver, error) {
	permit, err := r.Repos.GetCreateable(ctx, r.CreateableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	createable, ok := resolver.(createable)
	if !ok {
		return nil, errors.New("cannot convert resolver to createable")
	}
	return &createableResolver{createable}, nil
}

func (r *createdEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *createdEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *createdEventResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Event.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *createdEventResolver) User(ctx context.Context) (*userResolver, error) {
	userID, err := r.Event.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
