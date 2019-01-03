package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type removedFromActivityEventResolver struct {
	ActivityID *mytype.OID
	AssetID    *mytype.OID
	Conf       *myconf.Config
	Event      *repo.EventPermit
	Repos      *repo.Repos
}

func (r *removedFromActivityEventResolver) Activity(ctx context.Context) (*activityResolver, error) {
	activity, err := r.Repos.Activity().Get(ctx, r.ActivityID.String)
	if err != nil {
		return nil, err
	}
	return &activityResolver{Activity: activity, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *removedFromActivityEventResolver) Asset(ctx context.Context) (*userAssetResolver, error) {
	userAsset, err := r.Repos.UserAsset().Get(ctx, r.AssetID.String)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *removedFromActivityEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *removedFromActivityEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *removedFromActivityEventResolver) Study(ctx context.Context) (*studyResolver, error) {
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

func (r *removedFromActivityEventResolver) User(ctx context.Context) (*userResolver, error) {
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
