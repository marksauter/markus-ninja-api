package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type moveActivityAssetPayloadResolver struct {
	ActivityID *mytype.OID
	AssetID    *mytype.OID
	Conf       *myconf.Config
	Repos      *repo.Repos
}

func (r *moveActivityAssetPayloadResolver) Activity(
	ctx context.Context,
) (*activityResolver, error) {
	activity, err := r.Repos.Activity().Get(ctx, r.ActivityID.String)
	if err != nil {
		return nil, err
	}

	return &activityResolver{Activity: activity, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *moveActivityAssetPayloadResolver) AssetEdge(
	ctx context.Context,
) (*userAssetEdgeResolver, error) {
	userAsset, err := r.Repos.UserAsset().Get(ctx, r.AssetID.String)
	if err != nil {
		return nil, err
	}

	return NewUserAssetEdgeResolver(userAsset, r.Repos, r.Conf)
}
