package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteUserAssetPayloadResolver struct {
	Conf        *myconf.Config
	Repos       *repo.Repos
	StudyID     *mytype.OID
	UserAssetID *mytype.OID
}

func (r *deleteUserAssetPayloadResolver) DeletedUserAssetID() graphql.ID {
	return graphql.ID(r.UserAssetID.String)
}

func (r *deleteUserAssetPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
