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

type CoursePermit struct {
	checkFieldPermission FieldPermissionFunc
	course               *data.Course
}

func (r *CoursePermit) Get() *data.Course {
	course := r.course
	fields := structs.Fields(course)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return course
}

func (r *CoursePermit) AdvancedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("advanced_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if r.course.AdvancedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.course.AdvancedAt.Time, nil
}

func (r *CoursePermit) CompletedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("completed_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if r.course.CompletedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.course.CompletedAt.Time, nil
}

func (r *CoursePermit) AppledAt() time.Time {
	return r.course.AppledAt.Time
}

func (r *CoursePermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.course.CreatedAt.Time, nil
}

func (r *CoursePermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.course.Description.String, nil
}

func (r *CoursePermit) EnrolledAt() time.Time {
	return r.course.EnrolledAt.Time
}

func (r *CoursePermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.course.ID, nil
}

func (r *CoursePermit) IsPublished() (bool, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return r.course.PublishedAt.Status != pgtype.Null, nil
}

func (r *CoursePermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.course.Name.String, nil
}

func (r *CoursePermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.course.Number.Int, nil
}

func (r *CoursePermit) PublishedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if r.course.PublishedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.course.PublishedAt.Time, nil
}

func (r *CoursePermit) Status() (*mytype.CourseStatus, error) {
	if ok := r.checkFieldPermission("status"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.course.Status, nil
}

func (r *CoursePermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.course.StudyID, nil
}

func (r *CoursePermit) TopicedAt() time.Time {
	return r.course.TopicedAt.Time
}

func (r *CoursePermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.course.UpdatedAt.Time, nil
}

func (r *CoursePermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.course.UserID, nil
}

func NewCourseRepo(conf *myconf.Config) *CourseRepo {
	return &CourseRepo{
		conf: conf,
		load: loader.NewCourseLoader(),
	}
}

type CourseRepo struct {
	conf   *myconf.Config
	load   *loader.CourseLoader
	permit *Permitter
}

func (r *CourseRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	courses []*data.Course,
) ([]*CoursePermit, error) {
	coursePermits := make([]*CoursePermit, 0, len(courses))
	for _, l := range courses {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			coursePermits = append(coursePermits, &CoursePermit{fieldPermFn, l})
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *CourseRepo) Close() {
	r.load.ClearAll()
}

func (r *CourseRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *CourseRepo) CountByApplee(
	ctx context.Context,
	appleeID string,
	filters *data.CourseFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountCourseByApplee(db, appleeID, filters)
}

func (r *CourseRepo) CountByEnrollee(
	ctx context.Context,
	enrolleeID string,
	filters *data.CourseFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountCourseByEnrollee(db, enrolleeID, filters)
}

func (r *CourseRepo) CountByTopic(
	ctx context.Context,
	topicID string,
	filters *data.CourseFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountCourseByTopic(db, topicID, filters)
}

func (r *CourseRepo) CountBySearch(
	ctx context.Context,
	filters *data.CourseFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountCourseBySearch(db, filters)
}

func (r *CourseRepo) CountByStudy(
	ctx context.Context,
	studyID string,
	filters *data.CourseFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountCourseByStudy(db, studyID, filters)
}

func (r *CourseRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.CourseFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountCourseByUser(db, userID, filters)
}

func (r *CourseRepo) Create(
	ctx context.Context,
	c *data.Course,
) (*CoursePermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, c); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	course, err := data.CreateCourse(db, c)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) Get(
	ctx context.Context,
	id string,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	course, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) Pull(
	ctx context.Context,
	id string,
) (*CoursePermit, error) {
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
	course, err := data.GetCourse(db, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) GetByApplee(
	ctx context.Context,
	appleeID string,
	po *data.PageOptions,
	filters *data.CourseFilterOptions,
) ([]*CoursePermit, error) {
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
	courses, err := data.GetCourseByApplee(db, appleeID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, courses)
}

func (r *CourseRepo) GetByEnrollee(
	ctx context.Context,
	enrolleeID string,
	po *data.PageOptions,
	filters *data.CourseFilterOptions,
) ([]*CoursePermit, error) {
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
	courses, err := data.GetCourseByEnrollee(db, enrolleeID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, courses)
}

func (r *CourseRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
	filters *data.CourseFilterOptions,
) ([]*CoursePermit, error) {
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
	courses, err := data.GetCourseByStudy(db, studyID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, courses)
}

func (r *CourseRepo) GetByTopic(
	ctx context.Context,
	topicID string,
	po *data.PageOptions,
	filters *data.CourseFilterOptions,
) ([]*CoursePermit, error) {
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
	courses, err := data.GetCourseByTopic(db, topicID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, courses)
}

func (r *CourseRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.CourseFilterOptions,
) ([]*CoursePermit, error) {
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
	courses, err := data.GetCourseByUser(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, courses)
}

func (r *CourseRepo) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	course, err := r.load.GetByName(ctx, studyID, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) GetByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	course, err := r.load.GetByNumber(ctx, studyID, number)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) GetByStudyAndName(
	ctx context.Context,
	study,
	name string,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	course, err := r.load.GetByStudyAndName(ctx, study, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) Delete(
	ctx context.Context,
	course *data.Course,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, course); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteCourse(db, course.ID.String)
}

func (r *CourseRepo) IsPublishable(
	ctx context.Context,
	id string,
) (bool, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	return data.IsCoursePublishable(db, id)
}

func (r *CourseRepo) Search(
	ctx context.Context,
	po *data.PageOptions,
	filters *data.CourseFilterOptions,
) ([]*CoursePermit, error) {
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
	courses, err := data.SearchCourse(db, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, courses)
}

func (r *CourseRepo) Update(
	ctx context.Context,
	c *data.Course,
) (*CoursePermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, c); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	course, err := data.UpdateCourse(db, c)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) ViewerCanAdmin(
	ctx context.Context,
	c *data.Course,
) (bool, error) {
	return r.permit.ViewerCanAdmin(ctx, c)
}
