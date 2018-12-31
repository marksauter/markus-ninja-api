package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type removeActivityAssetPayloadResolver struct {
	ActivityID *mytype.OID
	AssetID    *mytype.OID
	Conf       *myconf.Config
	Repos      *repo.Repos
}

func (r *removeActivityAssetPayloadResolver) Activity(
	ctx context.Context,
) (*activityResolver, error) {
	activity, err := r.Repos.Activity().Get(ctx, r.ActivityID.String)
	if err != nil {
		return nil, err
	}

	return &activityResolver{Activity: activity, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *removeActivityAssetPayloadResolver) RemovedAssetID() graphql.ID {
	return graphql.ID(r.AssetID.String)
}

func (r *removeActivityAssetPayloadResolver) RemovedAssetEdge(
	ctx context.Context,
) (*userAssetEdgeResolver, error) {
	userAsset, err := r.Repos.UserAsset().Get(ctx, r.AssetID.String)
	if err != nil {
		return nil, err
	}

	return NewUserAssetEdgeResolver(userAsset, r.Repos, r.Conf)
}
