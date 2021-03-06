package resolver

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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

type lessonResolver struct {
	Conf   *myconf.Config
	Lesson *repo.LessonPermit
	Repos  *repo.Repos
}

func (r *lessonResolver) Activities(
	ctx context.Context,
	args struct {
		After    *string
		Before   *string
		FilterBy *data.ActivityFilterOptions
		First    *int32
		Last     *int32
		OrderBy  *OrderArg
	},
) (*activityConnectionResolver, error) {
	resolver := activityConnectionResolver{}
	lessonID, err := r.Lesson.ID()
	if err != nil {
		if err != repo.ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return &resolver, err
	}
	activityOrder, err := ParseActivityOrder(args.OrderBy)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	pageOptions, err := data.NewPageOptions(
		args.After,
		args.Before,
		args.First,
		args.Last,
		activityOrder,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}

	filters := data.ActivityFilterOptions{}
	if args.FilterBy != nil {
		filters = *args.FilterBy
	}

	activities, err := r.Repos.Activity().GetByLesson(
		ctx,
		lessonID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	activityConnectionResolver, err := NewActivityConnectionResolver(
		activities,
		pageOptions,
		lessonID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return &resolver, err
	}
	return activityConnectionResolver, nil
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
	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *lessonResolver) Body() (string, error) {
	body, err := r.Lesson.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *lessonResolver) BodyHTML(ctx context.Context) (mygql.HTML, error) {
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
) (*commentConnectionResolver, error) {
	lessonID, err := r.Lesson.ID()
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
		lessonID.String,
		pageOptions,
		&filters,
	)
	if err != nil {
		return nil, err
	}
	commentConnectionResolver, err := NewCommentConnectionResolver(
		comments,
		pageOptions,
		lessonID,
		&filters,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return commentConnectionResolver, nil
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
	return &courseResolver{Course: course, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *lessonResolver) CourseNumber() (*int32, error) {
	return r.Lesson.CourseNumber()
}

func (r *lessonResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Lesson.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonResolver) Draft() (string, error) {
	return r.Lesson.Draft()
}

func (r *lessonResolver) DraftBackup(
	ctx context.Context,
	args struct{ ID string },
) (*lessonDraftBackupResolver, error) {
	lessonID, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(args.ID, 10, 32)
	if err != nil {
		return nil, errors.New("invalid backup id")
	}

	draftBackup, err := r.Repos.LessonDraftBackup().Get(ctx, lessonID.String, int32(id))
	if err != nil {
		return nil, err
	}
	return &lessonDraftBackupResolver{
		Conf:              r.Conf,
		LessonDraftBackup: draftBackup,
		Repos:             r.Repos,
	}, nil
}

func (r *lessonResolver) DraftBackups(
	ctx context.Context,
) ([]*lessonDraftBackupResolver, error) {
	resolvers := []*lessonDraftBackupResolver{}

	lessonID, err := r.Lesson.ID()
	if err != nil {
		return resolvers, err
	}

	draftBackups, err := r.Repos.LessonDraftBackup().GetByLesson(ctx, lessonID.String)
	if err != nil {
		return resolvers, err
	}

	resolvers = make([]*lessonDraftBackupResolver, len(draftBackups))
	for i, b := range draftBackups {
		resolvers[i] = &lessonDraftBackupResolver{
			Conf:              r.Conf,
			LessonDraftBackup: b,
			Repos:             r.Repos,
		}
	}

	return resolvers, nil
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
		users,
		pageOptions,
		lessonID,
		args.FilterBy,
		r.Repos,
		r.Conf,
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

func (r *lessonResolver) IsPublished() (bool, error) {
	return r.Lesson.IsPublished()
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
	labelConnectionResolver, err := NewLabelConnectionResolver(
		labels,
		pageOptions,
		lessonID,
		args.FilterBy,
		r.Repos,
		r.Conf,
	)
	if err != nil {
		return nil, err
	}
	return labelConnectionResolver, nil
}

func (r *lessonResolver) LastEditedAt() (graphql.Time, error) {
	t, err := r.Lesson.LastEditedAt()
	return graphql.Time{t}, err
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
	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
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
	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}

func (r *lessonResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.Lesson.PublishedAt()
	if err != nil {
		return nil, err
	}
	return &graphql.Time{t}, nil
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
	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
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
		Types: &[]data.EventTypeFilter{
			data.EventTypeFilter{
				ActionIsNot: &actionIsNot,
				Type:        data.LessonEvent,
			},
		},
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

	resolver, err := NewLessonTimelineConnectionResolver(
		events,
		pageOptions,
		lessonID,
		filters,
		r.Repos,
		r.Conf,
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
	uri = mygql.URI(fmt.Sprintf("%s%s", r.Conf.ClientURL, resourcePath))
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

func (r *lessonResolver) ViewerNewComment(ctx context.Context) (*commentResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	lessonID, err := r.Lesson.ID()
	if err != nil {
		return nil, err
	}

	commentPermit, err := r.Repos.Comment().GetUserNewComment(
		ctx,
		viewer.ID.String,
		lessonID.String,
	)
	if err != nil {
		if err != data.ErrNotFound {
			return nil, err
		}
		studyID, err := r.Lesson.StudyID()
		if err != nil {
			return nil, err
		}
		comment := &data.Comment{}
		if err := comment.CommentableID.Set(lessonID); err != nil {
			mylog.Log.WithError(err).Error("failed to set comment commentable_id")
			return nil, myerr.SomethingWentWrongError
		}
		if err := comment.StudyID.Set(studyID); err != nil {
			mylog.Log.WithError(err).Error("failed to set comment user_id")
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
