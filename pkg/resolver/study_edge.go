package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyEdgeResolver(
	node *repo.StudyPermit,
	repos *repo.Repos,
) (*studyEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &studyEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
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
