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

func (r *LessonPermit) CourseID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("course_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.CourseID, nil
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
	return &r.lesson.ID, nil
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

func (r *LessonPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.StudyID, nil
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

func (r *LessonPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lesson.UserID, nil
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
	userID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByEnrollee(db, userID)
}

func (r *LessonRepo) CountByLabel(
	ctx context.Context,
	labelID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByLabel(db, labelID)
}

func (r *LessonRepo) CountBySearch(
	ctx context.Context,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonBySearch(db, query)
}

func (r *LessonRepo) CountByCourse(
	ctx context.Context,
	courseID string,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByCourse(db, courseID, filters)
}

func (r *LessonRepo) CountByStudy(
	ctx context.Context,
	studyID string,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByStudy(db, studyID, filters)
}

func (r *LessonRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonByUser(db, userID, filters)
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

func (r *LessonRepo) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	if err := r.CheckConnection(); err != nil {
		return false, err
	}
	return r.load.Exists(ctx, id)
}

func (r *LessonRepo) ExistsByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (bool, error) {
	if err := r.CheckConnection(); err != nil {
		return false, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return false, &myctx.ErrNotFound{"queryer"}
	}
	return data.ExistsLessonByNumber(db, studyID, number)
}

func (r *LessonRepo) ExistsByOwnerStudyAndNumber(
	ctx context.Context,
	ownerLogin,
	studyName string,
	number int32,
) (bool, error) {
	if err := r.CheckConnection(); err != nil {
		return false, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return false, &myctx.ErrNotFound{"queryer"}
	}
	return data.ExistsLessonByOwnerStudyAndNumber(db, ownerLogin, studyName, number)
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
	userID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByEnrollee(db, userID, po, filters)
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
	labelID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByLabel(db, labelID, po, filters)
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
	courseID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByCourse(db, courseID, po, filters)
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
	studyID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByStudy(db, studyID, po, filters)
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
	userID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessons, err := data.GetLessonByUser(db, userID, po, filters)
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
	courseID string,
	courseNumber int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lesson, err := data.GetLessonByCourseNumber(db, courseID, courseNumber)
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
	studyID string,
	number int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lesson, err := data.GetLessonByNumber(db, studyID, number)
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
	return data.DeleteLesson(db, lesson.ID.String)
}

func (r *LessonRepo) Search(
	ctx context.Context,
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
	lessons, err := data.SearchLesson(db, query, po)
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
