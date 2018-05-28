package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonCommentEdgeResolver(
	node *repo.LessonCommentPermit,
	repos *repo.Repos,
) (*lessonCommentEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &lessonCommentEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type lessonCommentEdgeResolver struct {
	cursor string
	node   *repo.LessonCommentPermit
	repos  *repo.Repos
}

func (r *lessonCommentEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonCommentEdgeResolver) Node() *lessonCommentResolver {
	return &lessonCommentResolver{LessonComment: r.node, Repos: r.repos}
}
