package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteStudyPayloadResolver struct {
	Conf    *myconf.Config
	OwnerID *mytype.OID
	Repos   *repo.Repos
	StudyID *mytype.OID
}

func (r *deleteStudyPayloadResolver) DeletedStudyID(
	ctx context.Context,
) graphql.ID {
	return graphql.ID(r.StudyID.String)
}

func (r *deleteStudyPayloadResolver) Owner(
	ctx context.Context,
) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.OwnerID.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
