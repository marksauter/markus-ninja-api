package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteStudyPayload = deleteStudyPayloadResolver

type deleteStudyPayloadResolver struct {
	OwnerId *mytype.OID
	StudyId *mytype.OID
	Repos   *repo.Repos
}

func (r *deleteStudyPayloadResolver) DeletedStudyId(
	ctx context.Context,
) graphql.ID {
	return graphql.ID(r.StudyId.String)
}

func (r *deleteStudyPayloadResolver) Owner(
	ctx context.Context,
) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.OwnerId.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}
