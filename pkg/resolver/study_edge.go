package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

func NewStudyEdgeResolver(
	cursor string,
	node *repo.StudyPermit,
	repos *repo.Repos,
) *studyEdgeResolver {
	return &studyEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type studyEdgeResolver struct {
	cursor string
	node   *repo.StudyPermit
	repos  *repo.Repos
}

func (r *studyEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *studyEdgeResolver) Node() *studyResolver {
	return &studyResolver{Study: r.node, Repos: r.repos}
}
