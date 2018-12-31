package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type addCommentPayloadResolver struct {
	Conf    *myconf.Config
	Comment *repo.CommentPermit
	Repos   *repo.Repos
}

func (r *addCommentPayloadResolver) CommentEdge() (*commentEdgeResolver, error) {
	return NewCommentEdgeResolver(r.Comment, r.Repos, r.Conf)
}

func (r *addCommentPayloadResolver) Commentable(
	ctx context.Context,
) (*commentableResolver, error) {
	commentableID, err := r.Comment.CommentableID()
	if err != nil {
		return nil, err
	}
	permit, err := r.Repos.GetCommentable(ctx, commentableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	commentable, ok := resolver.(commentable)
	if !ok {
		return nil, errors.New("cannot convert resolver to commentable")
	}
	return &commentableResolver{commentable}, nil
}
