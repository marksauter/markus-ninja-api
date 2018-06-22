package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrolledStudyEdgeResolver(
	cursor string,
	node *repo.StudyPermit,
	repos *repo.Repos,
) *enrolledStudyEdgeResolver {
	return &enrolledStudyEdgeResolver{
		cursor: cursor,
		node:   node,
		repos:  repos,
	}
}

type enrolledStudyEdgeResolver struct {
	cursor string
	node   *repo.StudyPermit
	repos  *repo.Repos
}

func (r *enrolledStudyEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *enrolledStudyEdgeResolver) EnrolledAt() (graphql.Time, error) {
	t, err := r.node.EnrolledAt()
	return graphql.Time{t}, err
}

func (r *enrolledStudyEdgeResolver) Node() *studyResolver {
	return &studyResolver{Study: r.node, Repos: r.repos}
}
