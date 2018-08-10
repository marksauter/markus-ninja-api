package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddCommentPayload = addCommentPayloadResolver

type addCommentPayloadResolver struct {
	Comment       *repo.CommentPermit
	CommentableId *mytype.OID
	Event         *repo.EventPermit
	Repos         *repo.Repos
}

func (r *addCommentPayloadResolver) Commentable(
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

func (r *addCommentPayloadResolver) CommentEdge() (*lessonCommentEdgeResolver, error) {
	return NewCommentEdgeResolver(r.Comment, r.Repos)
}

func (r *addCommentPayloadResolver) TimelineEdge() (*lessonTimelineEventEdgeResolver, error) {
	return NewTimelineEventEdgeResolver(r.Event, r.Repos)
}
