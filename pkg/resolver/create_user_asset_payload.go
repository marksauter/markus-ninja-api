package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createUserAssetPayloadResolver struct {
	UserAsset *repo.UserAssetPermit
	StudyId   *mytype.OID
	UserId    *mytype.OID
	Repos     *repo.Repos
}

func (r *createUserAssetPayloadResolver) UserAssetEdge() (*userAssetEdgeResolver, error) {
	return NewUserAssetEdgeResolver(r.UserAsset, r.Repos)
}

func (r *createUserAssetPayloadResolver) Owner(ctx context.Context) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.UserId.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *createUserAssetPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyId.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
