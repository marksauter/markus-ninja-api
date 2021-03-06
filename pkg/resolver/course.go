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
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type courseResolver struct {
	Conf   *myconf.Config
	Course *repo.CoursePermit
	Repos  *repo.Repos
}

func (r *courseResolver) AdvancedAt() (*graphql.Time, error) {
	t, err := r.Course.AdvancedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *courseResolver) CompletedAt() (*graphql.Time, error) {
	t, err := r.Course.CompletedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *courseResolver) AppleGivers(
	ctx context.Context,
	args AppleGiversArgs,
) (*appleGiverConnectionResolver, error) {
	appleGiverOrder, err := ParseAppleGiverOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		appleGiverOrder,
	)
	if err != nil {
		return nil, err
	}

	courseID, err := r.Course.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetByAppleable(
		ctx,
		courseID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	appleGiverConnectionResolver, err := NewAppleGiverConnectionResolver(
		users,
		pageOptions,
		courseID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return appleGiverConnectionResolver, nil
}

func (r *courseResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Course.CreatedAt()
	return graphql.Time{t}, err
}

func (r *courseResolver) Description() (string, error) {
	return r.Course.Description()
}

func (r *courseResolver) DescriptionHTML() (mygql.HTML, error) {
	description, err := r.Description()
	if err != nil {
		return "", err
	}
	descriptionHTML := util.MarkdownToHTML([]byte(description))
	gqlHTML := mygql.HTML(descriptionHTML)
	return gqlHTML, nil
}

func (r *courseResolver) ID() (graphql.ID, error) {
	id, err := r.Course.ID()
	return graphql.ID(id.String), err
}

func (r *courseResolver) IsPublished() (bool, error) {
	return r.Course.IsPublished()
}

func (r *courseResolver) IsPublishable(ctx context.Context) (bool, error) {
	isPublished, err := r.IsPublished()
	if err != nil {
		return false, err
	}
	if isPublished {
		return false, nil
	}
	id, err := r.Course.ID()
	if err != nil {
		return false, err
	}
	return r.Repos.Course().IsPublishable(ctx, id.String)
}

func (r *courseResolver) Lesson(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	courseID, err := r.Course.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().GetByCourseNumber(
		ctx,
		courseID.String,
		args.Number,
	)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *courseResolver) Lessons(
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
	courseID, err := r.Course.ID()
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
	lessons, err := r.Repos.Lesson().GetByCourse(
		ctx,
		courseID.String,
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
		courseID,
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

func (r *courseResolver) Name() (string, error) {
	return r.Course.Name()
}

func (r *courseResolver) Number() (int32, error) {
	return r.Course.Number()
}

func (r *courseResolver) Owner(ctx context.Context) (*userResolver, error) {
	userID, err := r.Course.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *courseResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.Course.PublishedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		return &graphql.Time{*t}, nil
	}
	return nil, nil
}

func (r *courseResolver) ResourcePath(
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
	uri = mygql.URI(fmt.Sprintf("%s/course/%d", string(studyPath), number))
	return uri, nil
}

func (r *courseResolver) Status() (string, error) {
	status, err := r.Course.Status()
	if err != nil {
		return "", err
	}
	return status.String(), nil
}

func (r *courseResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyID, err := r.Course.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *courseResolver) Topics(
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
	courseID, err := r.Course.ID()
	if err != nil {
		return nil, err
	}
	topicOrder, err := ParseTopicOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		topicOrder,
	)
	if err != nil {
		return nil, err
	}

	topics, err := r.Repos.Topic().GetByTopicable(
		ctx,
		courseID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	topicConnectionResolver, err := NewTopicConnectionResolver(
		topics,
		pageOptions,
		courseID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return topicConnectionResolver, nil
}

func (r *courseResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Course.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *courseResolver) URL(
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

func (r *courseResolver) ViewerCanAdmin(ctx context.Context) (bool, error) {
	course := r.Course.Get()
	return r.Repos.Course().ViewerCanAdmin(ctx, course)
}

func (r *courseResolver) ViewerCanApple(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	courseID, err := r.Course.ID()
	if err != nil {
		return false, err
	}

	appled := &data.Appled{}
	appled.AppleableID.Set(courseID)
	appled.UserID.Set(viewer.ID)
	return r.Repos.Appled().ViewerCanApple(ctx, appled)
}

func (r *courseResolver) ViewerHasAppled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	courseID, err := r.Course.ID()
	if err != nil {
		return false, err
	}

	appled := &data.Appled{}
	appled.AppleableID.Set(courseID)
	appled.UserID.Set(viewer.ID)
	if _, err := r.Repos.Appled().Get(ctx, appled); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
