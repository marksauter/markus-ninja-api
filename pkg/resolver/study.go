package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type Study = studyResolver

type studyResolver struct {
	Study *repo.StudyPermit
	Repos *repo.Repos
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

	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetByAppleable(
		ctx,
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.User().CountByAppleable(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	appleGiverConnectionResolver, err := NewAppleGiverConnectionResolver(
		users,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return appleGiverConnectionResolver, nil
}

func (r *studyResolver) Asset(
	ctx context.Context,
	args struct{ Name string },
) (*userAssetResolver, error) {
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAsset, err := r.Repos.UserAsset().GetByName(
		ctx,
		userId.String,
		studyId.String,
		args.Name,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Repos: r.Repos}, nil
}

func (r *studyResolver) Assets(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*userAssetConnectionResolver, error) {
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

	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	userAssets, err := r.Repos.UserAsset().GetByStudy(
		ctx,
		userId,
		studyId,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.UserAsset().CountByStudy(
		ctx,
		userId.String,
		studyId.String,
	)
	if err != nil {
		return nil, err
	}
	userAssetConnectionResolver, err := NewUserAssetConnectionResolver(
		userAssets,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return userAssetConnectionResolver, nil
}

func (r *studyResolver) Course(
	ctx context.Context,
	args struct{ Number int32 },
) (*courseResolver, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	course, err := r.Repos.Course().GetByNumber(
		ctx,
		studyId.String,
		args.Number,
	)
	if err != nil {
		return nil, err
	}
	return &courseResolver{Course: course, Repos: r.Repos}, nil
}

func (r *studyResolver) Courses(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*courseConnectionResolver, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	courseOrder, err := ParseCourseOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		courseOrder,
	)
	if err != nil {
		return nil, err
	}

	courses, err := r.Repos.Course().GetByStudy(
		ctx,
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Course().CountByStudy(
		ctx,
		studyId.String,
	)
	if err != nil {
		return nil, err
	}
	resolver, err := NewCourseConnectionResolver(
		courses,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
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

func (r *studyResolver) EnrolleeCount(ctx context.Context) (int32, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		var n int32
		return n, err
	}
	return r.Repos.User().CountByEnrollable(ctx, studyId.String)
}

func (r *studyResolver) Enrollees(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*enrolleeConnectionResolver, error) {
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

	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetByEnrollable(
		ctx,
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.User().CountByEnrollable(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	enrolleeConnectionResolver, err := NewEnrolleeConnectionResolver(
		users,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
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
	enrolled.EnrollableId.Set(id)
	enrolled.UserId.Set(viewer.Id)
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

func (r *studyResolver) Labels(
	ctx context.Context,
	args struct {
		After  *string
		Before *string
		First  *int32
		Last   *int32
		Query  *string
	},
) (*labelConnectionResolver, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		&LabelOrder{data.ASC, LabelName},
	)
	if err != nil {
		return nil, err
	}

	var count int32
	var labels []*repo.LabelPermit
	if args.Query != nil {
		count, err = r.Repos.Label().CountBySearch(ctx, studyId, *args.Query)
		if err != nil {
			return nil, err
		}
		labels, err = r.Repos.Label().Search(ctx, studyId, *args.Query, pageOptions)
		if err != nil {
			return nil, err
		}
	} else {
		count, err = r.Repos.Label().CountByStudy(ctx, studyId.String)
		if err != nil {
			return nil, err
		}
		labels, err = r.Repos.Label().GetByStudy(ctx, studyId.String, pageOptions)
		if err != nil {
			return nil, err
		}
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

func (r *studyResolver) Lesson(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().GetByNumber(
		ctx,
		studyId.String,
		args.Number,
	)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

func (r *studyResolver) Lessons(
	ctx context.Context,
	args struct {
		After          *string
		Before         *string
		First          *int32
		IsCourseLesson *bool
		Last           *int32
		OrderBy        *OrderArg
	},
) (*lessonConnectionResolver, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
	}
	lessonOrder, err := ParseLessonOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		lessonOrder,
	)
	if err != nil {
		return nil, err
	}

	filterOptions := []data.LessonFilterOption{}
	if args.IsCourseLesson != nil {
		if *args.IsCourseLesson {
			filterOptions = append(filterOptions, data.LessonIsCourseLesson)
		} else {
			filterOptions = append(filterOptions, data.LessonIsNotCourseLesson)
		}
	}

	lessons, err := r.Repos.Lesson().GetByStudy(
		ctx,
		studyId.String,
		pageOptions,
		filterOptions...,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Lesson().CountByStudy(
		ctx,
		studyId.String,
	)
	if err != nil {
		return nil, err
	}
	lessonConnectionResolver, err := NewLessonConnectionResolver(
		lessons,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
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
	studyId, err := r.Study.ID()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	lessonComments, err := r.Repos.LessonComment().GetByStudy(
		ctx,
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.LessonComment().CountByStudy(
		ctx,
		studyId.String,
	)
	if err != nil {
		return nil, err
	}
	lessonCommentConnectionResolver, err := NewLessonCommentConnectionResolver(
		lessonComments,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return lessonCommentConnectionResolver, nil
}

func (r *studyResolver) LessonCount(ctx context.Context) (int32, error) {
	studyId, err := r.Study.ID()
	if err != nil {
		var count int32
		return count, err
	}
	return r.Repos.Lesson().CountByStudy(
		ctx,
		studyId.String,
	)
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
	userId, err := r.Study.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *studyResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	nameWithOwner, err := r.NameWithOwner(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("/%s", nameWithOwner))
	return uri, nil
}

func (r *studyResolver) Topics(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*topicConnectionResolver, error) {
	studyId, err := r.Study.ID()
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
		studyId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Topic().CountByTopicable(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	topicConnectionResolver, err := NewTopicConnectionResolver(
		topics,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
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
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *studyResolver) ViewerCanAdmin(ctx context.Context) (bool, error) {
	study := r.Study.Get()
	return r.Repos.Study().ViewerCanAdmin(ctx, study)
}

func (r *studyResolver) ViewerCanEnroll(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableId.Set(studyId)
	enrolled.UserId.Set(viewer.Id)
	return r.Repos.Enrolled().ViewerCanEnroll(ctx, enrolled)
}

func (r *studyResolver) ViewerHasAppled(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	studyId, err := r.Study.ID()
	if err != nil {
		return false, err
	}

	appled := &data.Appled{}
	appled.AppleableId.Set(studyId)
	appled.UserId.Set(viewer.Id)
	if _, err := r.Repos.Appled().Get(ctx, appled); err != nil {
		if err == data.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
