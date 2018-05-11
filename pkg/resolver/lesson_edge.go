package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewLessonEdgeResolver(
	cursor string,
	node *repo.LessonPermit,
	repos *repo.Repos,
) *lessonEdgeResolver {
	return &lessonEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
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
