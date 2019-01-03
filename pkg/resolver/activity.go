package resolver

import (
	"context"
	"fmt"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type activityResolver struct {
	Conf     *myconf.Config
	Activity *repo.ActivityPermit
	Repos    *repo.Repos
}

func (r *activityResolver) AdvancedAt() (*graphql.Time, error) {
	t, err := r.Activity.AdvancedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *activityResolver) Asset(
	ctx context.Context,
	args struct{ Number int32 },
) (*userAssetResolver, error) {
	activityID, err := r.Activity.ID()
	if err != nil {
		return nil, err
	}
	userAsset, err := r.Repos.UserAsset().GetByActivityNumber(
		ctx,
		activityID.String,
		args.Number,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *activityResolver) Assets(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.UserAssetFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*userAssetConnectionResolver, error) {
	resolver := userAssetConnectionResolver{}
	activityID, err := r.Activity.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	userAssetOrder, err := ParseUserAssetOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		userAssetOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	filters := data.UserAssetFilterOptions{}
	if args.FilterBy != nil {
		filters = *args.FilterBy
	}

	userAssets, err := r.Repos.UserAsset().GetByActivity(
		ctx,
		activityID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	userAssetConnectionResolver, err := NewUserAssetConnectionResolver(
		userAssets,
		pageOptions,
		activityID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return userAssetConnectionResolver, nil
}

func (r *activityResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Activity.CreatedAt()
	return graphql.Time{t}, err
}

func (r *activityResolver) Description() (string, error) {
	return r.Activity.Description()
}

func (r *activityResolver) DescriptionHTML() (mygql.HTML, error) {
	description, err := r.Description()
	if err != nil {
		return "", err
	}
	descriptionHTML := util.MarkdownToHTML([]byte(description))
	gqlHTML := mygql.HTML(descriptionHTML)
	return gqlHTML, nil
}

func (r *activityResolver) ID() (graphql.ID, error) {
	id, err := r.Activity.ID()
	return graphql.ID(id.String), err
}

func (r *activityResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lessonID, err := r.Activity.LessonID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(ctx, lessonID.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *activityResolver) Name() (string, error) {
	return r.Activity.Name()
}

func (r *activityResolver) Number() (int32, error) {
	return r.Activity.Number()
}

func (r *activityResolver) Owner(ctx context.Context) (*userResolver, error) {
	userID, err := r.Activity.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *activityResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study(ctx)
	if err != nil {
		return uri, err
	}
	studyPath, err := study.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	number, err := r.Number()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/activity/%d", string(studyPath), number))
	return uri, nil
}

func (r *activityResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Activity.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *activityResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Activity.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *activityResolver) URL(
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

func (r *activityResolver) ViewerCanAdmin(ctx context.Context) (bool, error) {
	activity := r.Activity.Get()
	return r.Repos.Activity().ViewerCanAdmin(ctx, activity)
}
