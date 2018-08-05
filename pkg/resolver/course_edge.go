package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewCourseEdgeResolver(
	cursor string,
	node *repo.CoursePermit,
	repos *repo.Repos,
) *courseEdgeResolver {
	return &courseEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type courseEdgeResolver struct {
	cursor string
	node   *repo.CoursePermit
	repos  *repo.Repos
}

func (r *courseEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *courseEdgeResolver) Node() *courseResolver {
	return &courseResolver{Course: r.node, Repos: r.repos}
}
