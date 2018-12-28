package resolver

import (
	"context"
	"errors"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteCommentPayloadResolver struct {
	Conf          *myconf.Config
	CommentID     *mytype.OID
	CommentableID *mytype.OID
	Repos         *repo.Repos
}

func (r *deleteCommentPayloadResolver) DeletedCommentID() graphql.ID {
	return graphql.ID(r.CommentID.String)
}

func (r *deleteCommentPayloadResolver) Commentable(ctx context.Context) (*commentableResolver, error) {
	permit, err := r.Repos.GetCommentable(ctx, r.CommentableID)
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
