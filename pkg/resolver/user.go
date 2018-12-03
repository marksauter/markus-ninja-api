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

type userResolver struct {
	Conf  *myconf.Config
	Repos *repo.Repos
	User  *repo.UserPermit
}

func (r *userResolver) AccountUpdatedAt() (graphql.Time, error) {
	t, err := r.User.AccountUpdatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) Activity(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*userActivityConnectionResolver, error) {
	resolver := userActivityConnectionResolver{}

	filters := &data.EventFilterOptions{}
	ok, err := r.IsViewer(ctx)
	if err != nil && err != repo.ErrAccessDenied {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	} else if !ok {
		filters.IsPublic = util.NewBool(true)
	}

	userID, err := r.User.ID()
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
			Type: mytype.CourseEvent.String(),
		},
		data.EventTypeFilter{
			ActionIs: &[]string{mytype.PublishedAction.String()},
			Type:     mytype.LessonEvent.String(),
		},
		data.EventTypeFilter{
			Type: mytype.StudyEvent.String(),
		},
	}
	filters.Types = &eventTypes
	events, err := r.Repos.Event().GetByUser(
		ctx,
		userID.String,
		pageOptions,
		filters,
	)
	if err != nil {
		return &resolver, err
	}

	return NewUserActivityConnectionResolver(
		events,
		pageOptions,
		userID,
		filters,
		r.Repos,
		r.Conf,
	)
}

func (r *userResolver) Appled(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
		Search  *string
		Type    string
	},
) (*appleableConnectionResolver, error) {
	appleableType, err := ParseAppleableType(args.Type)
	if err != nil {
		return nil, err
	}
	appleableOrder, err := ParseAppleableOrder(appleableType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		appleableOrder,
	)
	if err != nil {
		return nil, err
	}

	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	permits := make([]repo.NodePermit, 0, pageOptions.Limit())
	switch appleableType {
	case AppleableTypeCourse:
		filters := &data.CourseFilterOptions{
			Search: args.Search,
		}
		courses, err := r.Repos.Course().GetByApplee(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		for _, s := range courses {
			permits = append(permits, s)
		}
	case AppleableTypeStudy:
		filters := &data.StudyFilterOptions{
			Search: args.Search,
		}
		studies, err := r.Repos.Study().GetByApplee(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		for _, s := range studies {
			permits = append(permits, s)
		}
	default:
		return nil, fmt.Errorf("invalid type %s for appleable type", appleableType.String())
	}

	return NewAppleableConnectionResolver(
		permits,
		pageOptions,
		id,
		args.Search,
		r.Repos,
		r.Conf,
	)
}

func (r *userResolver) Assets(
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
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	userAssetOrder, err := ParseUserAssetOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		userAssetOrder,
	)
	if err != nil {
		return nil, err
	}

	userAssets, err := r.Repos.UserAsset().GetByUser(
		ctx,
		id,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	resolver, err := NewUserAssetConnectionResolver(
		userAssets,
		pageOptions,
		id,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) Bio() (string, error) {
	return r.User.Bio()
}

func (r *userResolver) BioHTML() (mygql.HTML, error) {
	bio, err := r.Bio()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", bio))
	return h, nil
}

func (r *userResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.User.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) Courses(
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
	resolver := &courseConnectionResolver{}
	userID, err := r.User.ID()
	if err != nil {
		return resolver, err
	}
	courseOrder, err := ParseCourseOrder(args.OrderBy)
	if err != nil {
		return resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		courseOrder,
	)
	if err != nil {
		return resolver, err
	}

	filters := data.CourseFilterOptions{}
	if args.FilterBy != nil {
		filters = *args.FilterBy
	}

	ok, err := r.IsViewer(ctx)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return resolver, err
	} else if !ok {
		filters.IsPublished = util.NewBool(true)
	}
	courses, err := r.Repos.Course().GetByUser(
		ctx,
		userID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return resolver, err
	}
	resolver, err = NewCourseConnectionResolver(
		courses,
		pageOptions,
		userID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return resolver, err
	}
	return resolver, nil
}

func (r *userResolver) Email(ctx context.Context) (*emailResolver, error) {
	id, err := r.User.ProfileEmailID()
	if err != nil {
		return nil, err
	}

	email, err := r.Repos.Email().Get(ctx, id.String)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &emailResolver{Email: email, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userResolver) Emails(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.EmailFilterOptions
		First    *int32
		Last     *int32
	},
) (*emailConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	order := NewEmailOrder(data.ASC, EmailType)
	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		order,
	)
	if err != nil {
		return nil, err
	}

	if args.FilterBy != nil && args.FilterBy.Types != nil {
		for _, t := range *args.FilterBy.Types {
			_, err := mytype.ParseEmailType(t)
			if err != nil {
				return nil, err
			}
		}
	}

	tx, err, newTx := myctx.TransactionFromContext(ctx)
	if err != nil {
		return nil, err
	} else if newTx {
		defer data.RollbackTransaction(tx)
	}
	ctx = myctx.NewQueryerContext(ctx, tx)

	emails, err := r.Repos.Email().GetByUser(ctx, id.String, pageOptions, args.FilterBy)
	if err != nil {
		return nil, err
	}
	resolver, err := NewEmailConnectionResolver(
		emails,
		pageOptions,
		id,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) Enrolled(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
		Search  *string
		Type    string
	},
) (*enrollableConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	enrollableType, err := ParseEnrollableType(args.Type)
	if err != nil {
		return nil, err
	}
	enrollableOrder, err := ParseEnrollableOrder(enrollableType, args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		enrollableOrder,
	)
	if err != nil {
		return nil, err
	}

	permits := make([]repo.NodePermit, 0, pageOptions.Limit())

	switch enrollableType {
	case EnrollableTypeLesson:
		filters := &data.LessonFilterOptions{
			Search: args.Search,
		}
		lessons, err := r.Repos.Lesson().GetByEnrollee(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		for _, l := range lessons {
			permits = append(permits, l)
		}
	case EnrollableTypeStudy:
		filters := &data.StudyFilterOptions{
			Search: args.Search,
		}
		studies, err := r.Repos.Study().GetByEnrollee(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		for _, s := range studies {
			permits = append(permits, s)
		}
	case EnrollableTypeUser:
		filters := &data.UserFilterOptions{
			Search: args.Search,
		}
		users, err := r.Repos.User().GetByEnrollee(ctx, id.String, pageOptions, filters)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			permits = append(permits, u)
		}
	}
	return NewEnrollableConnectionResolver(
		permits,
		pageOptions,
		id,
		args.Search,
		r.Repos,
		r.Conf,
	)
}

func (r *userResolver) Enrollees(
	ctx context.Context,
	args EnrolleesArgs,
) (*enrolleeConnectionResolver, error) {
	id, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	enrolleeOrder, err := ParseEnrolleeOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		enrolleeOrder,
	)
	if err != nil {
		return nil, err
	}

	users, err := r.Repos.User().GetByEnrollable(
		ctx,
		id.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	resolver, err := NewEnrolleeConnectionResolver(
		users,
		pageOptions,
		id,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) EnrollmentStatus(ctx context.Context) (string, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return "", errors.New("viewer not found")
	}
	id, err := r.User.ID()
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

func (r *userResolver) Notifications(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*notificationConnectionResolver, error) {
	userID, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	notificationOrder, err := ParseNotificationOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		notificationOrder,
	)
	if err != nil {
		return nil, err
	}

	notifications, err := r.Repos.Notification().GetByUser(
		ctx,
		userID.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	notificationConnectionResolver, err := NewNotificationConnectionResolver(
		notifications,
		pageOptions,
		userID,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return notificationConnectionResolver, nil
}

func (r *userResolver) ID() (graphql.ID, error) {
	id, err := r.User.ID()
	return graphql.ID(id.String), err
}

func (r *userResolver) IsVerified(ctx context.Context) (bool, error) {
	return r.User.Verified()
}

func (r *userResolver) IsSiteAdmin() bool {
	for _, role := range r.User.Roles() {
		if role == data.AdminRole {
			return true
		}
	}
	return false
}

func (r *userResolver) IsViewer(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	id, err := r.User.ID()
	if err != nil {
		return false, err
	}
	return viewer.ID.String == id.String, nil
}

func (r *userResolver) Lessons(
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
	resolver := &lessonConnectionResolver{}
	userID, err := r.User.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return resolver, err
	}
	lessonOrder, err := ParseLessonOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return resolver, err
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
		return resolver, err
	}

	filters := data.LessonFilterOptions{}
	if args.FilterBy != nil {
		filters = *args.FilterBy
	}

	ok, err := r.IsViewer(ctx)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return resolver, err
	} else if !ok {
		filters.IsPublished = util.NewBool(true)
	}
	lessons, err := r.Repos.Lesson().GetByUser(
		ctx,
		userID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return resolver, err
	}
	resolver, err = NewLessonConnectionResolver(
		lessons,
		pageOptions,
		userID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return resolver, err
	}
	return resolver, nil
}

func (r *userResolver) Login() (string, error) {
	return r.User.Login()
}

func (r *userResolver) Name() (string, error) {
	return r.User.Name()
}

func (r *userResolver) ProfileUpdatedAt() (graphql.Time, error) {
	t, err := r.User.ProfileUpdatedAt()
	return graphql.Time{t}, err
}

func (r *userResolver) ReceivedActivity(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*userReceivedActivityConnectionResolver, error) {
	userID, err := r.User.ID()
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

	events, err := r.Repos.Event().GetReceivedByUser(
		ctx,
		userID.String,
		pageOptions,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return NewUserReceivedActivityConnectionResolver(
		events,
		pageOptions,
		userID,
		r.Repos,
		r.Conf,
	)
}

func (r *userResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	login, err := r.User.Login()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", login))
	return uri, nil
}

func (r *userResolver) Study(
	ctx context.Context,
	args struct{ Name string },
) (*studyResolver, error) {
	userID, err := r.User.ID()
	if err != nil {
		return nil, err
	}

	study, err := r.Repos.Study().GetByName(ctx, userID.String, args.Name)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *userResolver) Studies(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.StudyFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*studyConnectionResolver, error) {
	userID, err := r.User.ID()
	if err != nil {
		return nil, err
	}
	studyOrder, err := ParseStudyOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		studyOrder,
	)
	if err != nil {
		return nil, err
	}

	studies, err := r.Repos.Study().GetByUser(
		ctx,
		userID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	resolver, err := NewStudyConnectionResolver(
		studies,
		pageOptions,
		userID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *userResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", r.Conf.ClientURL, resourcePath))
	return uri, nil
}

func (r *userResolver) ViewerCanEnroll(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userID, err := r.User.ID()
	if err != nil {
		return false, err
	}

	if viewer.ID.String == userID.String {
		return false, err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableID.Set(userID)
	enrolled.UserID.Set(viewer.ID)
	return r.Repos.Enrolled().ViewerCanEnroll(ctx, enrolled)
}
