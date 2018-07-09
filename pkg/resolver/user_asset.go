package resolver

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type UserAsset = userAssetResolver

type userAssetResolver struct {
	UserAsset *repo.UserAssetPermit
	Repos     *repo.Repos
}

func (r *userAssetResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.UserAsset.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userAssetResolver) Href() (mygql.URI, error) {
	var uri mygql.URI
	href, err := r.UserAsset.Href()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(href)
	return uri, nil
}

func (r *userAssetResolver) ID() (graphql.ID, error) {
	id, err := r.UserAsset.ID()
	return graphql.ID(id.String), err
}

func (r *userAssetResolver) Name() (string, error) {
	return r.UserAsset.Name()
}

func (r *userAssetResolver) OriginalName() (string, error) {
	return r.UserAsset.OriginalName()
}

func (r *userAssetResolver) Owner(ctx context.Context) (*userResolver, error) {
	userId, err := r.UserAsset.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *userAssetResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.UserAsset.PublishedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *userAssetResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study()
	if err != nil {
		return uri, err
	}
	studyResourcePath, err := study.ResourcePath()
	if err != nil {
		return uri, err
	}
	name, err := r.UserAsset.Name()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/asset/%s", studyResourcePath, name))
	return uri, nil
}

func (r *userAssetResolver) Size() (int32, error) {
	s, err := r.UserAsset.Size()
	return int32(s), err
}

func (r *userAssetResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.UserAsset.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *userAssetResolver) Subtype() (string, error) {
	return r.UserAsset.Subtype()
}

func (r *userAssetResolver) Type() (string, error) {
	return r.UserAsset.Type()
}

func (r *userAssetResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.UserAsset.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *userAssetResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, resourcePath))
	return uri, nil
}

func (r *userAssetResolver) ViewerCanDelete(ctx context.Context) bool {
	userAsset := r.UserAsset.Get()
	return r.Repos.UserAsset().ViewerCanDelete(ctx, userAsset)
}

func (r *userAssetResolver) ViewerCanUpdate(ctx context.Context) bool {
	userAsset := r.UserAsset.Get()
	return r.Repos.UserAsset().ViewerCanUpdate(ctx, userAsset)
}
