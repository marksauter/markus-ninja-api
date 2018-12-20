package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewStudyEdgeResolver(
	node *repo.StudyPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*studyEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &studyEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type studyEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.StudyPermit
	repos  *repo.Repos
}

func (r *studyEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *studyEdgeResolver) Node() *studyResolver {
	return &studyResolver{Study: r.node, Conf: r.conf, Repos: r.repos}
}
