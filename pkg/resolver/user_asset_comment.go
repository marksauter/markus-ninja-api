package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type userAssetCommentResolver struct {
	UserAssetComment *repo.UserAssetCommentPermit
	Repos            *repo.Repos
}

func (r *userAssetCommentResolver) Asset(ctx context.Context) (*userAssetResolver, error) {
	userAssetId, err := r.UserAssetComment.AssetId()
	if err != nil {
		return nil, err
	}
	userAsset, err := r.Repos.UserAsset().Get(ctx, userAssetId.String)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Repos: r.Repos}, nil
}

func (r *userAssetCommentResolver) Author(ctx context.Context) (*userResolver, error) {
	userId, err := r.UserAssetComment.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *userAssetCommentResolver) Body() (string, error) {
	body, err := r.UserAssetComment.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *userAssetCommentResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.UserAssetComment.Body()
	if err != nil {
		return "", err
	}
	return mygql.HTML(body.ToHTML()), nil
}

func (r *userAssetCommentResolver) BodyText() (string, error) {
	body, err := r.UserAssetComment.Body()
	if err != nil {
		return "", err
	}
	return body.ToText(), nil
}

func (r *userAssetCommentResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.UserAssetComment.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userAssetCommentResolver) ID() (graphql.ID, error) {
	id, err := r.UserAssetComment.ID()
	return graphql.ID(id.String), err
}

func (r *userAssetCommentResolver) Labels(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*labelConnectionResolver, error) {
	userAssetCommentId, err := r.UserAssetComment.ID()
	if err != nil {
		return nil, err
	}
	labelOrder, err := ParseLabelOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		labelOrder,
	)
	if err != nil {
		return nil, err
	}

	labels, err := r.Repos.Label().GetByLabelable(
		ctx,
		userAssetCommentId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Label().CountByLabelable(ctx, userAssetCommentId.String)
	if err != nil {
		return nil, err
	}
	labelConnectionResolver, err := NewLabelConnectionResolver(
		labels,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return labelConnectionResolver, nil
}

func (r *userAssetCommentResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.UserAssetComment.PublishedAt()
	if err != nil {
		return nil, err
	}
	return &graphql.Time{t}, nil
}

func (r *userAssetCommentResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	userAsset, err := r.Asset(ctx)
	if err != nil {
		return uri, err
	}
	userAssetPath, err := userAsset.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	createdAt, err := r.UserAssetComment.CreatedAt()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf(
		"%s#asset-comment%d",
		string(userAssetPath),
		createdAt.Unix(),
	))
	return uri, nil
}

func (r *userAssetCommentResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.UserAssetComment.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *userAssetCommentResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.UserAssetComment.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *userAssetCommentResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *userAssetCommentResolver) ViewerCanDelete(ctx context.Context) bool {
	userAssetComment := r.UserAssetComment.Get()
	return r.Repos.UserAssetComment().ViewerCanDelete(ctx, userAssetComment)
}

func (r *userAssetCommentResolver) ViewerCanUpdate(ctx context.Context) bool {
	userAssetComment := r.UserAssetComment.Get()
	return r.Repos.UserAssetComment().ViewerCanUpdate(ctx, userAssetComment)
}

func (r *userAssetCommentResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userId, err := r.UserAssetComment.UserId()
	if err != nil {
		return false, err
	}

	return viewer.Id.String == userId.String, nil
}
