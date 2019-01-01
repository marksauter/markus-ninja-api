package resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/pgtype"
	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type userAssetResolver struct {
	Conf      *myconf.Config
	Repos     *repo.Repos
	UserAsset *repo.UserAssetPermit
}

func (r *userAssetResolver) Activity(ctx context.Context) (*activityResolver, error) {
	activityID, err := r.UserAsset.ActivityID()
	if err != nil {
		return nil, err
	}
	activity, err := r.Repos.Activity().Get(ctx, activityID.String)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &activityResolver{Activity: activity, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userAssetResolver) ActivityNumber() (*int32, error) {
	return r.UserAsset.ActivityNumber()
}

func (r *userAssetResolver) Comments(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*commentConnectionResolver, error) {
	userAssetID, err := r.UserAsset.ID()
	if err != nil {
		return nil, err
	}
	commentOrder, err := ParseCommentOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		commentOrder,
	)
	if err != nil {
		return nil, err
	}

	filters := data.CommentFilterOptions{
		IsPublished: util.NewBool(true),
	}
	comments, err := r.Repos.Comment().GetByCommentable(
		ctx,
		userAssetID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		return nil, err
	}
	commentConnectionResolver, err := NewCommentConnectionResolver(
		comments,
		pageOptions,
		userAssetID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return commentConnectionResolver, nil
}

func (r *userAssetResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.UserAsset.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userAssetResolver) Description() (string, error) {
	return r.UserAsset.Description()
}

func (r *userAssetResolver) DescriptionHTML() (mygql.HTML, error) {
	description, err := r.Description()
	if err != nil {
		return "", err
	}
	descriptionHTML := util.MarkdownToHTML([]byte(description))
	gqlHTML := mygql.HTML(descriptionHTML)
	return gqlHTML, nil
}

func (r *userAssetResolver) Href() (mygql.URI, error) {
	var uri mygql.URI
	key, err := r.UserAsset.Key()
	if err != nil {
		return uri, err
	}
	userID, err := r.UserAsset.UserID()
	if err != nil {
		return uri, err
	}
	href := fmt.Sprintf(
		r.Conf.ImagesURL+"/%s/%s",
		userID.Short,
		key,
	)
	uri = mygql.URI(href)
	return uri, nil
}

func (r *userAssetResolver) ID() (graphql.ID, error) {
	id, err := r.UserAsset.ID()
	return graphql.ID(id.String), err
}

func (r *userAssetResolver) IsActivityAsset() (bool, error) {
	activityID, err := r.UserAsset.ActivityID()
	if err != nil {
		return false, err
	}

	return activityID.Status != pgtype.Null, nil
}

func (r *userAssetResolver) Labels(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.LabelFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*labelConnectionResolver, error) {
	userAssetID, err := r.UserAsset.ID()
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
		userAssetID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	labelConnectionResolver, err := NewLabelConnectionResolver(
		labels,
		pageOptions,
		userAssetID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return labelConnectionResolver, nil
}

func (r *userAssetResolver) Name() (string, error) {
	return r.UserAsset.Name()
}

func (r *userAssetResolver) NextAsset(ctx context.Context) (*userAssetResolver, error) {
	activityID, err := r.UserAsset.ActivityID()
	if err != nil {
		return nil, err
	}
	activityNumber, err := r.UserAsset.ActivityNumber()
	if err != nil {
		return nil, err
	}
	if activityNumber == nil {
		return nil, nil
	}
	userAsset, err := r.Repos.UserAsset().GetByActivityNumber(
		ctx,
		activityID.String,
		*activityNumber+1,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userAssetResolver) OriginalName() (string, error) {
	return r.UserAsset.OriginalName()
}

func (r *userAssetResolver) Owner(ctx context.Context) (*userResolver, error) {
	userID, err := r.UserAsset.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userAssetResolver) PreviousAsset(ctx context.Context) (*userAssetResolver, error) {
	activityID, err := r.UserAsset.ActivityID()
	if err != nil {
		return nil, err
	}
	activityNumber, err := r.UserAsset.ActivityNumber()
	if err != nil {
		return nil, err
	}
	if activityNumber == nil || *activityNumber <= 1 {
		return nil, nil
	}
	userAsset, err := r.Repos.UserAsset().GetByActivityNumber(
		ctx,
		activityID.String,
		*activityNumber-1,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userAssetResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study(ctx)
	if err != nil {
		return uri, err
	}
	studyResourcePath, err := study.ResourcePath(ctx)
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
	studyID, err := r.UserAsset.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userAssetResolver) Subtype() (string, error) {
	return r.UserAsset.Subtype()
}

func (r *userAssetResolver) Timeline(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*userAssetTimelineConnectionResolver, error) {
	id, err := r.UserAsset.ID()
	if err != nil {
		return nil, err
	}
	eventOrder, err := ParseEventOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		eventOrder,
	)
	if err != nil {
		return nil, err
	}

	actionIsNot := []string{
		mytype.CreatedAction.String(),
	}
	filters := &data.EventFilterOptions{
		Types: &[]data.EventTypeFilter{
			data.EventTypeFilter{
				ActionIsNot: &actionIsNot,
				Type:        mytype.UserAssetEvent.String(),
			},
		},
	}
	events, err := r.Repos.Event().GetByUserAsset(
		ctx,
		id.String,
		pageOptions,
		filters,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := NewUserAssetTimelineConnectionResolver(
		events,
		pageOptions,
		id,
		filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userAssetResolver) Type() (string, error) {
	return r.UserAsset.Type()
}

func (r *userAssetResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.UserAsset.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *userAssetResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", r.Conf.ClientURL, resourcePath))
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

func (r *userAssetResolver) ViewerNewComment(ctx context.Context) (*commentResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	userAssetID, err := r.UserAsset.ID()
	if err != nil {
		return nil, err
	}

	commentPermit, err := r.Repos.Comment().GetUserNewComment(
		ctx,
		viewer.ID.String,
		userAssetID.String,
	)
	if err != nil {
		if err != data.ErrNotFound {
			return nil, err
		}
		studyID, err := r.UserAsset.StudyID()
		if err != nil {
			return nil, err
		}
		comment := &data.Comment{}
		if err := comment.CommentableID.Set(userAssetID); err != nil {
			mylog.Log.WithError(err).Error("failed to set comment commentable_id")
			return nil, myerr.SomethingWentWrongError
		}
		if err := comment.StudyID.Set(studyID); err != nil {
			mylog.Log.WithError(err).Error("failed to set comment user_id")
			return nil, myerr.SomethingWentWrongError
		}
		if err := comment.Type.Set(mytype.CommentableTypeUserAsset); err != nil {
			mylog.Log.WithError(err).Error("failed to set comment type")
			return nil, myerr.SomethingWentWrongError
		}
		if err := comment.UserID.Set(&viewer.ID); err != nil {
			mylog.Log.WithError(err).Error("failed to set comment user_id")
			return nil, myerr.SomethingWentWrongError
		}
		commentPermit, err = r.Repos.Comment().Create(ctx, comment)
		if err != nil {
			return nil, err
		}
	}

	return &commentResolver{Comment: commentPermit, Conf: r.Conf, Repos: r.Repos}, nil
}
