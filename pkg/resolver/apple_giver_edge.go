package resolver

import (
	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func NewAppleGiverEdgeResolver(
	node *repo.UserPermit,
	repos *repo.Repos,
	conf *myconf.Config,
) (*appleGiverEdgeResolver, error) {
	id, err := node.ID()
	if err != nil {
		return nil, err
	}
	cursor := data.EncodeCursor(id.String)
	return &appleGiverEdgeResolver{
		conf:   conf,
		cursor: cursor,
		node:   node,
		repos:  repos,
	}, nil
}

type appleGiverEdgeResolver struct {
	conf   *myconf.Config
	cursor string
	node   *repo.UserPermit
	repos  *repo.Repos
}

func (r *appleGiverEdgeResolver) AppledAt() graphql.Time {
	return graphql.Time{r.node.AppledAt()}
}

func (r *appleGiverEdgeResolver) Cursor() string {
	return r.cursor
}

func (r *appleGiverEdgeResolver) Node() *userResolver {
	return &userResolver{User: r.node, Conf: r.conf, Repos: r.repos}
}
