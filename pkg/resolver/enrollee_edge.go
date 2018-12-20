package resolver

import (
	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewEnrolleeEdgeResolver(
	node *repo.UserPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*enrolleeEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor, err := data.EncodeCursor(id.String)
	if err != nil {
		return nil, err
	}
	return &enrolleeEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type enrolleeEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *enrolleeEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *enrolleeEdgeResolver) EnrolledAt() graphql.Time {
	return graphql.Time{r.node.EnrolledAt()}
}

func (r *enrolleeEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Conf: r.conf, Repos: r.repos}
}
