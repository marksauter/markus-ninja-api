package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteEmailPayloadResolver struct {
	Conf    *myconf.Config
	EmailID *mytype.OID
	UserID  *mytype.OID
	Repos   *repo.Repos
}

func (r *deleteEmailPayloadResolver) DeletedEmailID() graphql.ID {
	return graphql.ID(r.EmailID.String)
}

func (r *deleteEmailPayloadResolver) User(ctx context.Context) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.UserID.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
