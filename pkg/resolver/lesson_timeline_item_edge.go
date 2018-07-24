package resolver

import (
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonTimelineItemEdgeResolver(
	node repo.NodePermit,
	repos *repo.Repos,
) (*lessonTimelineItemEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &lessonTimelineItemEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type lessonTimelineItemEdgeResolver struct {
	cursor string
	node   repo.NodePermit
	repos  *repo.Repos
}

func (r *lessonTimelineItemEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonTimelineItemEdgeResolver) Node() (*lessonTimelineItemResolver, error) {
	resolver, err := nodePermitToResolver(r.node, r.repos)
	if err != nil {
		return nil, err
	}
	item, ok := resolver.(lessonTimelineItem)
	if !ok {
		return nil, errors.New("cannot convert resolver to lesson_timeline_item")
	}
	return &lessonTimelineItemResolver{item}, nil
}
