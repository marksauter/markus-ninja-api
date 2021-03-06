package repo

import (
	"context"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
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
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lesson.Body, nil
}

func (r *LessonPermit) CourseID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("course_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lesson.CourseID, nil
}

func (r *LessonPermit) CourseNumber() (*int32, error) {
	if ok := r.checkFieldPermission("course_number"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if r.lesson.CourseNumber.Status == pgtype.Null {
		return nil, nil
	}
	return &r.lesson.CourseNumber.Int, nil
}

func (r *LessonPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lesson.CreatedAt.Time, nil
}

func (r *LessonPermit) Draft() (string, error) {
	if ok := r.checkFieldPermission("draft"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.lesson.Draft.String, nil
}

func (r *LessonPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lesson.ID, nil
}

func (r *LessonPermit) IsPublished() (bool, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return r.lesson.PublishedAt.Status != pgtype.Null, nil
}

func (r *LessonPermit) LastEditedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("last_edited_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lesson.LastEditedAt.Time, nil
}

func (r *LessonPermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.lesson.Number.Int, nil
}

func (r *LessonPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lesson.PublishedAt.Time, nil
}

func (r *LessonPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lesson.StudyID, nil
}

func (r *LessonPermit) Title() (string, error) {
	if ok := r.checkFieldPermission("title"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.lesson.Title.String, nil
}

func (r *LessonPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lesson.UpdatedAt.Time, nil
}

func (r *LessonPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lesson.UserID, nil
}

func NewLessonRepo(conf *myconf.Config) *LessonRepo {
	return &LessonRepo{
		conf: conf,
		load: loader.NewLessonLoader(),
	}
}

type LessonRepo struct {
	conf   *myconf.Config
	load   *loader.LessonLoader
	permit *Permitter
}

func (r *LessonRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	lessons []*data.Lesson,
) ([]*LessonPermit, error) {
	lessonPermits := make([]*LessonPermit, 0, len(lessons))
	for _, l := range lessons {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			lessonPermits = append(lessonPermits, &LessonPermit{fieldPermFn, l})
		}
	}
	return lessonPermits, nil
}

func (r *LessonRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *LessonRepo) Clear(id string) {
	r.load.Clear(id)
}

func (r *LessonRepo) Close() {
	r.load.ClearAll()
}

func (r *LessonRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *LessonRepo) CountByEnrollee(
	ctx context.Context,
	userID string,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLessonByEnrollee(db, userID, filters)
}

func (r *LessonRepo) CountByLabel(
	ctx context.Context,
	labelID string,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLessonByLabel(db, labelID, filters)
}

func (r *LessonRepo) CountBySearch(
	ctx context.Context,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLessonBySearch(db, filters)
}

func (r *LessonRepo) CountByCourse(
	ctx context.Context,
	courseID string,
	filters *data.LessonFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
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
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
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
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountLessonByUser(db, userID, filters)
}

func (r *LessonRepo) Create(
	ctx context.Context,
	l *data.Lesson,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, l); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := data.CreateLesson(db, l)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
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
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return data.ExistsLessonByOwnerStudyAndNumber(db, ownerLogin, studyName, number)
}

func (r *LessonRepo) Get(
	ctx context.Context,
	id string,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

// Same as Get(), but doesn't use the dataloader
func (r *LessonRepo) Pull(
	ctx context.Context,
	id string,
) (*LessonPermit, error) {
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := data.GetLesson(db, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.GetLessonByEnrollee(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) GetByLabel(
	ctx context.Context,
	labelID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.GetLessonByLabel(db, labelID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) GetByCourse(
	ctx context.Context,
	courseID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.GetLessonByCourse(db, courseID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.GetLessonByStudy(db, studyID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.GetLessonByUser(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) GetByCourseNumber(
	ctx context.Context,
	courseID string,
	courseNumber int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := data.GetLessonByCourseNumber(db, courseID, courseNumber)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := data.GetLessonByNumber(db, studyID, number)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) BatchGetByNumber(
	ctx context.Context,
	studyID string,
	numbers []int32,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.BatchGetLessonByNumber(db, studyID, numbers)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) GetByOwnerStudyAndNumber(
	ctx context.Context,
	owner,
	study string,
	number int32,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := data.GetLessonByOwnerStudyAndNumber(db, owner, study, number)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) Delete(
	ctx context.Context,
	lesson *data.Lesson,
) error {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, lesson); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteLesson(db, lesson.ID.String)
}

func (r *LessonRepo) Search(
	ctx context.Context,
	po *data.PageOptions,
	filters *data.LessonFilterOptions,
) ([]*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessons, err := data.SearchLesson(db, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonRepo) Update(
	ctx context.Context,
	l *data.Lesson,
) (*LessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := data.UpdateLesson(db, l)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonPermit{fieldPermFn, lesson}, nil
}

func (r *LessonRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.Lesson,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		if err != ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return false
	}
	return true
}

func (r *LessonRepo) ViewerCanUpdate(
	ctx context.Context,
	l *data.Lesson,
) bool {
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		if err == ErrFieldAccessDenied {
			return true
		} else if err != ErrAccessDenied {
			mylog.Log.WithError(err).Error(util.Trace(""))
		}
		return false
	}
	return true
}
