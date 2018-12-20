package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonEdgeResolver(
	node *repo.LessonPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*lessonEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &lessonEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type lessonEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.LessonPermit
	repos  *repo.Repos
}

func (r *lessonEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonEdgeResolver) Node() *lessonResolver {
	return &lessonResolver{Lesson: r.node, Conf: r.conf, Repos: r.repos}
}
