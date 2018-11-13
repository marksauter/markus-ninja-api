package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createUserAssetPayloadResolver struct {
	Conf      *myconf.Config
	Repos     *repo.Repos
	StudyID   *mytype.OID
	UserAsset *repo.UserAssetPermit
	UserID    *mytype.OID
}

func (r *createUserAssetPayloadResolver) UserAssetEdge() (*userAssetEdgeResolver, error) {
	return NewUserAssetEdgeResolver(r.UserAsset, r.Repos, r.Conf)
}

func (r *createUserAssetPayloadResolver) Owner(ctx context.Context) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.UserID.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *createUserAssetPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
