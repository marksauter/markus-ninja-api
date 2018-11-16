package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewLessonCommentEdgeResolver(
	node *repo.LessonCommentPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*lessonCommentEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &lessonCommentEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type lessonCommentEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.LessonCommentPermit
	repos  *repo.Repos
}

func (r *lessonCommentEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *lessonCommentEdgeResolver) Node() *lessonCommentResolver {
	return &lessonCommentResolver{LessonComment: r.node, Conf: r.conf, Repos: r.repos}
}
