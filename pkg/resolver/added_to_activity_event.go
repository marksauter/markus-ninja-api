package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type addedToActivityEventResolver struct {
	ActivityID *mytype.OID
	AssetID    *mytype.OID
	Conf       *myconf.Config
	Event      *repo.EventPermit
	Repos      *repo.Repos
}

func (r *addedToActivityEventResolver) Activity(ctx context.Context) (*activityResolver, error) {
	activity, err := r.Repos.Activity().Get(ctx, r.ActivityID.String)
	if err != nil {
		return nil, err
	}
	return &activityResolver{Activity: activity, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *addedToActivityEventResolver) Asset(ctx context.Context) (*userAssetResolver, error) {
	userAsset, err := r.Repos.UserAsset().Get(ctx, r.AssetID.String)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *addedToActivityEventResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Event.CreatedAt()
	return graphql.Time{t}, err
}

func (r *addedToActivityEventResolver) ID() (graphql.ID, error) {
	id, err := r.Event.ID()
	return graphql.ID(id.String), err
}

func (r *addedToActivityEventResolver) Study(ctx context.Context) (*studyResolver, error) {
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

func (r *addedToActivityEventResolver) User(ctx context.Context) (*userResolver, error) {
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
