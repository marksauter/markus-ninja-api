package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewUserAssetCommentEdgeResolver(
	node *repo.UserAssetCommentPermit,
	repos *repo.Repos,
) (*userAssetCommentEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &userAssetCommentEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type userAssetCommentEdgeResolver struct {
	cursor string
	node   *repo.UserAssetCommentPermit
	repos  *repo.Repos
}

func (r *userAssetCommentEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *userAssetCommentEdgeResolver) Node() *userAssetCommentResolver {
	return &userAssetCommentResolver{UserAssetComment: r.node, Repos: r.repos}
}
