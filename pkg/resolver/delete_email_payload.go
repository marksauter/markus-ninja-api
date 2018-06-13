package resolver

import (
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

func (r *deleteEmailPayloadResolver) User() (*userResolver, error) {
	user, err := r.Repos.User().Get(r.UserId.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}
