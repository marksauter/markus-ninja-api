package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type studyResolver struct {
	Conf  *myconf.Config
	Repos *repo.Repos
	Study *repo.StudyPermit
}

func (r *studyResolver) Activity(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*studyActivityConnectionResolver, error) {
	resolver := studyActivityConnectionResolver{}

	filters := &data.EventFilterOptions{}
	ok, err := r.ViewerCanAdmin(ctx)
	if err != nil && err != repo.ErrAccessDenied {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	} else if !ok {
		filters.IsPublic = util.NewBool(true)
	}

	studyID, err := r.Study.ID()
	if err != nil {
		return &resolver, err
	}
	eventOrder, err := ParseEventOrder(args.OrderBy)
	if err != nil {
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		eventOrder,
	)
	if err != nil {
		return &resolver, err
	}

	eventTypes := []data.EventTypeFilter{
		data.EventTypeFilter{
			ActionIs: &[]string{
				mytype.CreatedAction.String(),
			},
			Type: mytype.CourseEvent.String(),
		},
		data.EventTypeFilter{
			ActionIs: &[]string{
				mytype.CreatedAction.String(),
				mytype.PublishedAction.String(),
			},
			Type: mytype.LessonEvent.String(),
		},
	}
	filters.Types = &eventTypes
	events, err := r.Repos.Event().GetByStudy(
		ctx,
		studyID.String,
		pageOptions,
		filters,
	)
	if err != nil {
		return &resolver, err
	}

	return NewStudyActivityConnectionResolver(
		events,
		pageOptions,
		studyID,
		filters,
		r.Repos,
		r.Conf,
	)
}

func (r *studyResolver) AdvancedAt() (*graphql.Time, error) {
	t, err := r.Study.AdvancedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *studyResolver) AppleGivers(
	ctx context.Context,
	args AppleGiversArgs,
) (*appleGiverConnectionResolver, error) {
	resolver := appleGiverConnectionResolver{}
	appleGiverOrder, err := ParseAppleGiverOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		appleGiverOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	users, err := r.Repos.User().GetByAppleable(
		ctx,
		studyID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	appleGiverConnectionResolver, err := NewAppleGiverConnectionResolver(
		users,
		pageOptions,
		studyID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return appleGiverConnectionResolver, nil
}

func (r *studyResolver) Asset(
	ctx context.Context,
	args struct{ Name string },
) (*userAssetResolver, error) {
	studyID, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAsset, err := r.Repos.UserAsset().GetByName(
		ctx,
		studyID.String,
		args.Name,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{
		Conf:      r.Conf,
		Repos:     r.Repos,
		UserAsset: userAsset,
	}, nil
}

func (r *studyResolver) Assets(
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

	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	userAssets, err := r.Repos.UserAsset().GetByStudy(
		ctx,
		studyID,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	userAssetConnectionResolver, err := NewUserAssetConnectionResolver(
		userAssets,
		pageOptions,
		studyID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return userAssetConnectionResolver, nil
}

func (r *studyResolver) Course(
	ctx context.Context,
	args struct{ Number int32 },
) (*courseResolver, error) {
	studyID, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	course, err := r.Repos.Course().GetByNumber(
		ctx,
		studyID.String,
		args.Number,
	)
	if err != nil {
		return nil, err
	}
	return &courseResolver{
		Conf:   r.Conf,
		Course: course,
		Repos:  r.Repos,
	}, nil
}

func (r *studyResolver) Courses(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.CourseFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*courseConnectionResolver, error) {
	resolver := courseConnectionResolver{}
	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	courseOrder, err := ParseCourseOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		courseOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	filters := data.CourseFilterOptions{}
	if args.FilterBy != nil {
		filters = *args.FilterBy
	}

	ok, err := r.ViewerCanAdmin(ctx)
	if err != nil && err != repo.ErrAccessDenied {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	} else if !ok {
		filters.IsPublished = util.NewBool(true)
	}
	courses, err := r.Repos.Course().GetByStudy(
		ctx,
		studyID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	courseConnectionResolver, err := NewCourseConnectionResolver(
		courses,
		pageOptions,
		studyID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return courseConnectionResolver, nil
}

func (r *studyResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Study.CreatedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) Description() (string, error) {
	return r.Study.Description()
}

func (r *studyResolver) DescriptionHTML() (mygql.HTML, error) {
	description, err := r.Description()
	if err != nil {
		return "", err
	}
	descriptionHTML := util.MarkdownToHTML([]byte(description))
	gqlHTML := mygql.HTML(descriptionHTML)
	return gqlHTML, nil
}

func (r *studyResolver) Enrollees(
	ctx context.Context,
	args EnrolleesArgs,
) (*enrolleeConnectionResolver, error) {
	resolver := enrolleeConnectionResolver{}
	enrolleeOrder, err := ParseEnrolleeOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		enrolleeOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	studyID, err := r.Study.ID()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	users, err := r.Repos.User().GetByEnrollable(
		ctx,
		studyID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	enrolleeConnectionResolver, err := NewEnrolleeConnectionResolver(
		users,
		pageOptions,
		studyID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return enrolleeConnectionResolver, nil
}

func (r *studyResolver) EnrollmentStatus(ctx context.Context) (string, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return "", errors.New("viewer not found")
	}
	id, err := r.Study.ID()
	if err != nil {
		return "", err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableID.Set(id)
	enrolled.UserID.Set(viewer.ID)
	permit, err := r.Repos.Enrolled().Get(ctx, enrolled)
	if err != nil {
		if err != data.ErrNotFound {
			return "", err
		}
		return mytype.EnrollmentStatusUnenrolled.String(), nil
	}

	status, err := permit.Status()
	if err != nil {
		return "", err
	}
	return status.String(), nil
}

func (r *studyResolver) ID() (graphql.ID, error) {
	id, err := r.Study.ID()
	return graphql.ID(id.String), err
}

func (r *studyResolver) IsPrivate(ctx context.Context) (bool, error) {
	return r.Study.Private()
}

func (r *studyResolver) Label(
	ctx context.Context,
	args struct{ Name string },
) (*labelResolver, error) {
	studyID, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	label, err := r.Repos.Label().GetByName(
		ctx,
		studyID.String,
		args.Name,
	)
	if err != nil {
		return nil, err
	}
	return &labelResolver{
		Conf:  r.Conf,
		Label: label,
		Repos: r.Repos,
	}, nil
}

func (r *studyResolver) Labels(
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
	resolver := labelConnectionResolver{}
	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	labelOrder, err := ParseLabelOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		labelOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	labels, err := r.Repos.Label().GetByStudy(ctx, studyID.String, pageOptions, args.FilterBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	labelConnectionResolver, err := NewLabelConnectionResolver(
		labels,
		pageOptions,
		studyID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return labelConnectionResolver, nil
}

func (r *studyResolver) Lesson(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	studyID, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().GetByNumber(
		ctx,
		studyID.String,
		args.Number,
	)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{
		Conf:   r.Conf,
		Lesson: lesson,
		Repos:  r.Repos,
	}, nil
}

func (r *studyResolver) Lessons(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.LessonFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*lessonConnectionResolver, error) {
	resolver := lessonConnectionResolver{}
	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	lessonOrder, err := ParseLessonOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	filters := data.LessonFilterOptions{}
	if args.FilterBy != nil {
		filters = *args.FilterBy
	}

	ok, err := r.ViewerCanAdmin(ctx)
	if err != nil && err != repo.ErrAccessDenied {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	} else if !ok {
		filters.IsPublished = util.NewBool(true)
	}
	lessons, err := r.Repos.Lesson().GetByStudy(
		ctx,
		studyID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	lessonConnectionResolver, err := NewLessonConnectionResolver(
		lessons,
		pageOptions,
		studyID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return lessonConnectionResolver, nil
}

func (r *studyResolver) LessonComments(
	ctx context.Context,
	args struct {
		After  *string
		Before *string
		First  *int32
		Last   *int32
	},
) (*lessonCommentConnectionResolver, error) {
	resolver := lessonCommentConnectionResolver{}
	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	lessonCommentOrder := NewLessonCommentOrder(data.DESC, LessonCommentCreatedAt)

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonCommentOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	filters := data.LessonCommentFilterOptions{}
	ok, err := r.ViewerCanAdmin(ctx)
	if err != repo.ErrAccessDenied {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	} else if !ok {
		filters.IsPublished = util.NewBool(true)
	}
	lessonComments, err := r.Repos.LessonComment().GetByStudy(
		ctx,
		studyID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	lessonCommentConnectionResolver, err := NewLessonCommentConnectionResolver(
		lessonComments,
		pageOptions,
		studyID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return lessonCommentConnectionResolver, nil
}

func (r *studyResolver) Name() (string, error) {
	return r.Study.Name()
}

func (r *studyResolver) NameWithOwner(
	ctx context.Context,
) (string, error) {
	name, err := r.Name()
	if err != nil {
		return "", err
	}
	owner, err := r.Owner(ctx)
	if err != nil {
		return "", err
	}
	ownerLogin, err := owner.Login()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", ownerLogin, name), nil
}

func (r *studyResolver) Owner(ctx context.Context) (*userResolver, error) {
	userID, err := r.Study.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{
		Conf:  r.Conf,
		Repos: r.Repos,
		User:  user,
	}, nil
}

func (r *studyResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	owner, err := r.Owner(ctx)
	if err != nil {
		return uri, err
	}
	ownerResourcePath, err := owner.ResourcePath()
	if err != nil {
		return uri, err
	}
	name, err := r.Name()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", ownerResourcePath, name))
	return uri, nil
}

func (r *studyResolver) Topics(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.TopicFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*topicConnectionResolver, error) {
	resolver := topicConnectionResolver{}
	studyID, err := r.Study.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	topicOrder, err := ParseTopicOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		topicOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	topics, err := r.Repos.Topic().GetByTopicable(
		ctx,
		studyID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	topicConnectionResolver, err := NewTopicConnectionResolver(
		topics,
		pageOptions,
		studyID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return topicConnectionResolver, nil
}

func (r *studyResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Study.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *studyResolver) URL(
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

func (r *studyResolver) ViewerCanAdmin(ctx context.Context) (bool, error) {
	study := r.Study.Get()
	return r.Repos.Study().ViewerCanAdmin(ctx, study)
}

func (r *studyResolver) ViewerCanApple(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyID, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	appled := &data.Appled{}
	appled.AppleableID.Set(studyID)
	appled.UserID.Set(viewer.ID)
	return r.Repos.Appled().ViewerCanApple(ctx, appled)
}

func (r *studyResolver) ViewerCanEnroll(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyID, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableID.Set(studyID)
	enrolled.UserID.Set(viewer.ID)
	return r.Repos.Enrolled().ViewerCanEnroll(ctx, enrolled)
}

func (r *studyResolver) ViewerHasAppled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyID, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	appled := &data.Appled{}
	appled.AppleableID.Set(studyID)
	appled.UserID.Set(viewer.ID)
	if _, err := r.Repos.Appled().Get(ctx, appled); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
