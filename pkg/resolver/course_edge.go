package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewCourseEdgeResolver(
	node *repo.CoursePermit,
	repos *repo.Repos,
) (*courseEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &courseEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
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
