package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonEdgeResolver(
	node *repo.LessonPermit,
	repos *repo.Repos,
) (*lessonEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &lessonEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type lessonEdgeResolver struct {
	cursor string
	node   *repo.LessonPermit
	repos  *repo.Repos
}

func (r *lessonEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonEdgeResolver) Node() *lessonResolver {
	return &lessonResolver{Lesson: r.node, Repos: r.repos}
}
