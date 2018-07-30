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
)

type lessonResolver struct {
	Lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Author(ctx context.Context) (*userResolver, error) {
	userId, err := r.Lesson.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonResolver) Body() (string, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *lessonResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return mygql.HTML(body.ToHTML()), nil
}

func (r *lessonResolver) BodyText() (string, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return body.ToText(), nil
}

func (r *lessonResolver) Comments(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*lessonCommentConnectionResolver, error) {
	lessonId, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}
	lessonCommentOrder, err := ParseLessonCommentOrder(args.OrderBy)
	if err != nil {
		return nil, err
	}

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

	lessonComments, err := r.Repos.LessonComment().GetByLesson(
		ctx,
		lessonId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.LessonComment().CountByLesson(
		ctx,
		lessonId.String,
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

func (r *lessonResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Lesson.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) EnrolleeCount(ctx context.Context) (int32, error) {
	lessonId, err := r.Lesson.ID()
	if err != nil {
		var n int32
		return n, err
	}
	return r.Repos.User().CountByEnrollable(ctx, lessonId.String)
}

func (r *lessonResolver) Enrollees(
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

	lessonId, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetEnrollees(
		ctx,
		lessonId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.User().CountByEnrollable(ctx, lessonId.String)
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

func (r *lessonResolver) EnrollmentStatus(ctx context.Context) (string, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return "", errors.New("viewer not found")
	}
	id, err := r.Lesson.ID()
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

func (r *lessonResolver) HasNextLesson(ctx context.Context) (bool, error) {
	studyId, err := r.Lesson.StudyId()
	if err != nil {
		return false, err
	}
	count, err := r.Repos.Lesson().CountByStudy(ctx, studyId.String)
	if err != nil {
		return false, err
	}
	number, err := r.Lesson.Number()
	if err != nil {
		return false, err
	}
	return number < count, nil
}

func (r *lessonResolver) HasPrevLesson(ctx context.Context) (bool, error) {
	number, err := r.Lesson.Number()
	if err != nil {
		return false, err
	}
	return number > 1, nil
}

func (r *lessonResolver) ID() (graphql.ID, error) {
	id, err := r.Lesson.ID()
	return graphql.ID(id.String), err
}

func (r *lessonResolver) Labels(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*labelConnectionResolver, error) {
	lessonId, err := r.Lesson.ID()
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
		lessonId.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Label().CountByLabelable(ctx, lessonId.String)
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

func (r *lessonResolver) Number() (int32, error) {
	return r.Lesson.Number()
}

func (r *lessonResolver) PublishedAt() (graphql.Time, error) {
	t, err := r.Lesson.PublishedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) ResourcePath(
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
	uri = mygql.URI(fmt.Sprintf("%s/lesson/%d", string(studyPath), number))
	return uri, nil
}

func (r *lessonResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.Lesson.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonResolver) Timeline(
	ctx context.Context,
	args struct {
		After   *string
		Before  *string
		First   *int32
		Last    *int32
		OrderBy *OrderArg
	},
) (*lessonTimelineConnectionResolver, error) {
	lessonId, err := r.Lesson.ID()
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

	events, err := r.Repos.Event().GetByTarget(
		ctx,
		lessonId.String,
		pageOptions,
		data.FilterCreateEvents,
		data.FilterDismissEvents,
		data.FilterEnrollEvents,
	)
	if err != nil {
		return nil, err
	}

	count, err := r.Repos.Event().CountByTarget(
		ctx,
		lessonId.String,
		data.FilterCreateEvents,
		data.FilterDismissEvents,
		data.FilterEnrollEvents,
	)
	if err != nil {
		return nil, err
	}
	resolver, err := NewLessonTimelineConnectionResolver(
		events,
		pageOptions,
		count,
		r.Repos,
	)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *lessonResolver) Title() (string, error) {
	return r.Lesson.Title()
}

func (r *lessonResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.Lesson.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) URL(
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

func (r *lessonResolver) ViewerCanDelete(ctx context.Context) bool {
	lesson := r.Lesson.Get()
	return r.Repos.Lesson().ViewerCanDelete(ctx, lesson)
}

func (r *lessonResolver) ViewerCanUpdate(ctx context.Context) bool {
	lesson := r.Lesson.Get()
	return r.Repos.Lesson().ViewerCanUpdate(ctx, lesson)
}

func (r *lessonResolver) ViewerCanEnroll(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	lessonId, err := r.Lesson.ID()
	if err != nil {
		return false, err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableId.Set(lessonId)
	enrolled.UserId.Set(viewer.Id)
	return r.Repos.Enrolled().ViewerCanEnroll(ctx, enrolled)
}

func (r *lessonResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userId, err := r.Lesson.UserId()
	if err != nil {
		return false, err
	}

	return viewer.Id.String == userId.String, nil
}
