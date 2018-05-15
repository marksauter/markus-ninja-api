package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewLessonCommentEdgeResolver(
	cursor string,
	node *repo.LessonCommentPermit,
	repos *repo.Repos,
) *lessonCommentEdgeResolver {
	return &lessonCommentEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
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
