package resolver

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type lessonResolver struct {
	Lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Author(ctx context.Context) (*userResolver, error) {
	userID, err := r.Lesson.UserID()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userID.String)
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

func (r *lessonResolver) BodyHTML(ctx context.Context) (mygql.HTML, error) {
	lessonBody, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	body := *lessonBody
	studyID, err := r.Lesson.StudyID()
	if err != nil {
		return "", err
	}
	bodyStr := body.String

	study, err := r.Study(ctx)
	if err != nil {
		return "", err
	}
	studyPath, err := study.ResourcePath(ctx)
	if err != nil {
		return "", err
	}
	lessonNumberRefToLink := func(s string) string {
		result := mytype.NumberRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		number := result[1]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := r.Repos.Lesson().ExistsByNumber(ctx, studyID.String, int32(n))
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[#%[3]d](%[1]s%[2]s/lesson/%[3]d)",
			clientURL,
			studyPath,
			n,
		))
	}
	bodyStr = mytype.NumberRefRegexp.ReplaceAllStringFunc(bodyStr, lessonNumberRefToLink)

	crossStudyRefToLink := func(s string) string {
		result := mytype.CrossStudyRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		owner := result[1]
		name := result[2]
		number := result[3]
		n, err := strconv.ParseInt(number, 10, 32)
		if err != nil {
			return s
		}
		exists, err := r.Repos.Lesson().ExistsByOwnerStudyAndNumber(ctx, owner, name, int32(n))
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[%[2]s/%[3]s#%[4]d](%[1]s/%[2]s/%[3]s/lesson/%[4]d)",
			clientURL,
			owner,
			name,
			n,
		))
	}
	bodyStr = mytype.CrossStudyRefRegexp.ReplaceAllStringFunc(bodyStr, crossStudyRefToLink)

	userRefToLink := func(s string) string {
		result := mytype.AtRefRegexp.FindStringSubmatch(s)
		if len(result) == 0 {
			return s
		}
		name := result[1]
		exists, err := r.Repos.User().ExistsByLogin(ctx, name)
		if err != nil {
			return s
		}
		if !exists {
			return s
		}
		return util.ReplaceWithPadding(s, fmt.Sprintf("[@%[2]s](%[1]s/%[2]s)",
			clientURL,
			name,
		))
	}
	bodyStr = mytype.AtRefRegexp.ReplaceAllStringFunc(bodyStr, userRefToLink)

	if err := body.Set(bodyStr); err != nil {
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
	lessonID, err := r.Lesson.ID()
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
		lessonID.String,
		pageOptions,
	)
	if err != nil {
		return nil, err
	}
	lessonCommentConnectionResolver, err := NewLessonCommentConnectionResolver(
		r.Repos,
		lessonComments,
		pageOptions,
		lessonID,
	)
	if err != nil {
		return nil, err
	}
	return lessonCommentConnectionResolver, nil
}

func (r *lessonResolver) Course(ctx context.Context) (*courseResolver, error) {
	courseID, err := r.Lesson.CourseID()
	if err != nil {
		return nil, err
	}
	course, err := r.Repos.Course().Get(ctx, courseID.String)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &courseResolver{Course: course, Repos: r.Repos}, nil
}

func (r *lessonResolver) CourseNumber() (*int32, error) {
	return r.Lesson.CourseNumber()
}

func (r *lessonResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Lesson.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) Enrollees(
	ctx context.Context,
	args EnrolleesArgs,
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

	lessonID, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}
	users, err := r.Repos.User().GetByEnrollable(
		ctx,
		lessonID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	enrolleeConnectionResolver, err := NewEnrolleeConnectionResolver(
		r.Repos,
		users,
		pageOptions,
		lessonID,
		args.FilterBy,
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

func (r *lessonResolver) ID() (graphql.ID, error) {
	id, err := r.Lesson.ID()
	return graphql.ID(id.String), err
}

func (r *lessonResolver) IsCourseLesson() (bool, error) {
	courseID, err := r.Lesson.CourseID()
	if err != nil {
		return false, err
	}

	return courseID.Status != pgtype.Null, nil
}

func (r *lessonResolver) Labels(
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
	lessonID, err := r.Lesson.ID()
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
		lessonID.String,
		pageOptions,
		args.FilterBy,
	)
	if err != nil {
		return nil, err
	}
	count, err := r.Repos.Label().CountByLabelable(ctx, lessonID.String, args.FilterBy)
	if err != nil {
		return nil, err
	}
	labelConnectionResolver, err := NewLabelConnectionResolver(
		r.Repos,
		labels,
		pageOptions,
		count,
	)
	if err != nil {
		return nil, err
	}
	return labelConnectionResolver, nil
}

func (r *lessonResolver) Number() (int32, error) {
	return r.Lesson.Number()
}

func (r *lessonResolver) NextLesson(ctx context.Context) (*lessonResolver, error) {
	courseID, err := r.Lesson.CourseID()
	if err != nil {
		return nil, err
	}
	courseNumber, err := r.Lesson.CourseNumber()
	if err != nil {
		return nil, err
	}
	if courseNumber == nil {
		return nil, nil
	}
	lesson, err := r.Repos.Lesson().GetByCourseNumber(
		ctx,
		courseID.String,
		*courseNumber+1,
	)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

func (r *lessonResolver) PreviousLesson(ctx context.Context) (*lessonResolver, error) {
	courseID, err := r.Lesson.CourseID()
	if err != nil {
		return nil, err
	}
	courseNumber, err := r.Lesson.CourseNumber()
	if err != nil {
		return nil, err
	}
	if courseNumber == nil || *courseNumber <= 1 {
		return nil, nil
	}
	lesson, err := r.Repos.Lesson().GetByCourseNumber(
		ctx,
		courseID.String,
		*courseNumber-1,
	)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
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
	studyID, err := r.Lesson.StudyID()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyID.String)
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
	lessonID, err := r.Lesson.ID()
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
		mytype.MentionedAction.String(),
	}
	filters := &data.EventFilterOptions{
		ActionIsNot: &actionIsNot,
	}
	events, err := r.Repos.Event().GetByLesson(
		ctx,
		lessonID.String,
		pageOptions,
		filters,
	)
	if err != nil {
		return nil, err
	}

	count, err := r.Repos.Event().CountByLesson(
		ctx,
		lessonID.String,
		filters,
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
	lessonID, err := r.Lesson.ID()
	if err != nil {
		return false, err
	}

	enrolled := &data.Enrolled{}
	enrolled.EnrollableID.Set(lessonID)
	enrolled.UserID.Set(viewer.ID)
	return r.Repos.Enrolled().ViewerCanEnroll(ctx, enrolled)
}

func (r *lessonResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userID, err := r.Lesson.UserID()
	if err != nil {
		return false, err
	}

	return viewer.ID.String == userID.String, nil
}
