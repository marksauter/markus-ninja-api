package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type unlabeledEventResolver struct {
	Conf        *myconf.Config
	Event       *repo.EventPermit
	LabelID     *mytype.OID
	LabelableID *mytype.OID
	Repos       *repo.Repos
}

func (r *unlabeledEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *unlabeledEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *unlabeledEventResolver) Label(ctx context.Context) (*labelResolver, error) {
	unlabel, err := r.Repos.Label().Get(ctx, r.LabelID.String)
	if err != nil {
		return nil, err
	}
	return &labelResolver{Label: unlabel, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *unlabeledEventResolver) Labelable(ctx context.Context) (*labelableResolver, error) {
	permit, err := r.Repos.GetLabelable(ctx, r.LabelableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	labelable, ok := resolver.(labelable)
	if !ok {
		return nil, errors.New("cannot convert resolver to labelable")
	}
	return &labelableResolver{labelable}, nil
}

func (r *unlabeledEventResolver) Study(ctx context.Context) (*studyResolver, error) {
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

func (r *unlabeledEventResolver) User(ctx context.Context) (*userResolver, error) {
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
