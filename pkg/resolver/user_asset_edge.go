package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetEdgeResolver(
	node *repo.UserAssetPermit,
	repos *repo.Repos,
) (*userAssetEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &userAssetEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type userAssetEdgeResolver struct {
	cursor string
	node   *repo.UserAssetPermit
	repos  *repo.Repos
}

func (r *userAssetEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userAssetEdgeResolver) Node() *userAssetResolver {
	return &userAssetResolver{UserAsset: r.node, Repos: r.repos}
}
