package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetEdgeResolver(
	node *repo.UserAssetPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*userAssetEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &userAssetEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type userAssetEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.UserAssetPermit
	repos  *repo.Repos
}

func (r *userAssetEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userAssetEdgeResolver) Node() *userAssetResolver {
	return &userAssetResolver{UserAsset: r.node, Conf: r.conf, Repos: r.repos}
}
