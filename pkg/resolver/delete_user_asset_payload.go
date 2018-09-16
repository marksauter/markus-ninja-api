package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteUserAssetPayload = deleteUserAssetPayloadResolver

type deleteUserAssetPayloadResolver struct {
	UserAssetID *mytype.OID
	StudyID     *mytype.OID
	Repos       *repo.Repos
}

func (r *deleteUserAssetPayloadResolver) DeletedUserAssetID() graphql.ID {
	return graphql.ID(r.UserAssetID.String)
}

func (r *deleteUserAssetPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
