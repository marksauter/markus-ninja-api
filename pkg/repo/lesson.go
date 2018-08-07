package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type LessonPermit struct {
	checkFieldPermission FieldPermissionFunc
	lesson               *data.Lesson
}

func (r *LessonPermit) Get() *data.Lesson {
	lesson := r.lesson
	fields := structs.Fields(lesson)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return lesson
}

func (r *LessonPermit) Body() (*mytype.Markdown, error) {
	if ok := r.checkFieldPermission("body"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.Body, nil
}

func (r *LessonPermit) CourseId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("course_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.CourseId, nil
}

func (r *LessonPermit) CourseNumber() (*int32, error) {
	if ok := r.checkFieldPermission("course_number"); !ok {
		return nil, ErrAccessDenied
	}
	if r.lesson.CourseNumber.Status == pgtype.Null {
		return nil, nil
	}
	return &r.lesson.CourseNumber.Int, nil
}

func (r *LessonPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.CreatedAt.Time, nil
}

func (r *LessonPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.Id, nil
}

func (r *LessonPermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		var i int32
		return i, ErrAccessDenied
	}
	return r.lesson.Number.Int, nil
}

func (r *LessonPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.PublishedAt.Time, nil
}

func (r *LessonPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.StudyId, nil
}

func (r *LessonPermit) Title() (string, error) {
	if ok := r.checkFieldPermission("title"); !ok {
		return "", ErrAccessDenied
	}
	return r.lesson.Title.String, nil
}

func (r *LessonPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lesson.UpdatedAt.Time, nil
}

func (r *LessonPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.UserId, nil
}

func NewLessonRepo() *LessonRepo {
	return &LessonRepo{
		load: loader.NewLessonLoader(),
	}
}

type LessonRepo struct {
	load   *loader.LessonLoader
	permit *Permitter
}

func (r *LessonRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *LessonRepo) Close() {
	r.load.ClearAll()
}

func (r *LessonRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("lesson connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LessonRepo) CountByEnrollee(
	ctx context.Context,
	userId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByEnrollee(db, userId)
}

func (r *LessonRepo) CountByLabel(
	ctx context.Context,
	labelId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByLabel(db, labelId)
}

func (r *LessonRepo) CountBySearch(
	ctx context.Context,
	within *mytype.OID,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonBySearch(db, within, query)
}

func (r *LessonRepo) CountByCourse(
	ctx context.Context,
	courseId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByCourse(db, courseId)
}

func (r *LessonRepo) CountByStudy(
	ctx context.Context,
	studyId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByStudy(db, studyId)
}

func (r *LessonRepo) CountByUser(
	ctx context.Context,
	userId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByUser(db, userId)
}

func (r *LessonRepo) Create(
	ctx context.Context,
	l *data.Lesson,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, l); err != nil {
		return nil, err
	}
	lesson, err := data.CreateLesson(db, l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Get(
	ctx context.Context,
	id string,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lesson, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) GetByEnrollee(
	ctx context.Context,
	userId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByEnrollee(db, userId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByLabel(
	ctx context.Context,
	labelId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByLabel(db, labelId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByCourse(
	ctx context.Context,
	courseId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByCourse(db, courseId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByStudy(
	ctx context.Context,
	studyId string,
	po *data.PageOptions,
	opts ...data.LessonFilterOption,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByStudy(db, studyId, po, opts...)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByUser(
	ctx context.Context,
	userId string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByUser(db, userId, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) GetByCourseNumber(
	ctx context.Context,
	courseId string,
	courseNumber int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lesson, err := data.GetLessonByCourseNumber(db, courseId, courseNumber)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) GetByNumber(
	ctx context.Context,
	studyId string,
	number int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lesson, err := data.GetLessonByNumber(db, studyId, number)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Delete(
	ctx context.Context,
	lesson *data.Lesson,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, lesson); err != nil {
		return err
	}
	return data.DeleteLesson(db, lesson.Id.String)
}

func (r *LessonRepo) Search(
	ctx context.Context,
	within *mytype.OID,
	query string,
	po *data.PageOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.SearchLesson(db, within, query, po)
	if err != nil {
		return nil, err
	}
	lessonPermits := make([]*LessonPermit, len(lessons))
	if len(lessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessons {
			lessonPermits[i] = &LessonPermit{fieldPermFn, l}
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) Update(
	ctx context.Context,
	l *data.Lesson,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return nil, err
	}
	lesson, err := data.UpdateLesson(db, l)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.Lesson,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *LessonRepo) ViewerCanUpdate(
	ctx context.Context,
	l *data.Lesson,
) bool {
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
