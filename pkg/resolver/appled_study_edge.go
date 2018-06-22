package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppledStudyEdgeResolver(
	cursor string,
	node *repo.StudyPermit,
	repos *repo.Repos,
) *appledStudyEdgeResolver {
	return &appledStudyEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type appledStudyEdgeResolver struct {
	cursor string
	node   *repo.StudyPermit
	repos  *repo.Repos
}

func (r *appledStudyEdgeResolver) AppledAt() (graphql.Time, error) {
	t, err := r.node.AppledAt()
	return graphql.Time{t}, err
}

func (r *appledStudyEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *appledStudyEdgeResolver) Node() *studyResolver {
	return &studyResolver{Study: r.node, Repos: r.repos}
}
