package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteEmailPayload = deleteEmailPayloadResolver

type deleteEmailPayloadResolver struct {
	EmailId *mytype.OID
	UserId  *mytype.OID
	Repos   *repo.Repos
}

func (r *deleteEmailPayloadResolver) DeletedEmailId() graphql.ID {
	return graphql.ID(r.EmailId.String)
}

func (r *deleteEmailPayloadResolver) User(ctx context.Context) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.UserId.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}
