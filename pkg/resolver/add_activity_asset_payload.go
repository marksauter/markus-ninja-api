package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type addActivityAssetPayloadResolver struct {
	Conf       *myconf.Config
	ActivityID *mytype.OID
	AssetID    *mytype.OID
	Repos      *repo.Repos
}

func (r *addActivityAssetPayloadResolver) Activity(
	ctx context.Context,
) (*activityResolver, error) {
	activity, err := r.Repos.Activity().Get(ctx, r.ActivityID.String)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return &activityResolver{Activity: activity, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *addActivityAssetPayloadResolver) AssetEdge(
	ctx context.Context,
) (*userAssetEdgeResolver, error) {
	userAsset, err := r.Repos.UserAsset().Get(ctx, r.AssetID.String)
	if err != nil {
		return nil, err
	}

	return NewUserAssetEdgeResolver(userAsset, r.Repos, r.Conf)
}
