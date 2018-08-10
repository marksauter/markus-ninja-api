package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteCommentPayload = deleteCommentPayloadResolver

type deleteCommentPayloadResolver struct {
	CommentId       *mytype.OID
	TimelineEventId *mytype.OID
	Id              *mytype.OID
	Repos           *repo.Repos
}

func (r *deleteCommentPayloadResolver) DeletedCommentId() graphql.ID {
	return graphql.ID(r.CommentId.String)
}

func (r *deleteCommentPayloadResolver) DeletedTimelineEventId() graphql.ID {
	return graphql.ID(r.TimelineEventId.String)
}

func (r *deleteCommentPayloadResolver) Commentable(
	ctx context.Context,
) (*commentableResolver, error) {
	permit, err := r.Repos.GetCommentable(ctx, r.CommentableId)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	commentable, ok := resolver.(commentable)
	if !ok {
		return nil, errors.New("cannot convert resolver to commentable")
	}
	return &commentableResolver{commentable}, nil
}
