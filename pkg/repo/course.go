package repo

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
		return nil, ErrAccessDenied
	}
	if r.course.AdvancedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.course.AdvancedAt.Time, nil
}

func (r *CoursePermit) CompletedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("completed_at"); !ok {
		return nil, ErrAccessDenied
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
		return time.Time{}, ErrAccessDenied
	}
	return r.course.CreatedAt.Time, nil
}

func (r *CoursePermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		return "", ErrAccessDenied
	}
	return r.course.Description.String, nil
}

func (r *CoursePermit) EnrolledAt() time.Time {
	return r.course.EnrolledAt.Time
}

func (r *CoursePermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.course.Id, nil
}

func (r *CoursePermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.course.Name.String, nil
}

func (r *CoursePermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		var i int32
		return i, ErrAccessDenied
	}
	return r.course.Number.Int, nil
}

func (r *CoursePermit) Status() (*mytype.CourseStatus, error) {
	if ok := r.checkFieldPermission("status"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.course.Status, nil
}

func (r *CoursePermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.course.StudyId, nil
}

func (r *CoursePermit) TopicedAt() time.Time {
	return r.course.TopicedAt.Time
}

func (r *CoursePermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.course.UpdatedAt.Time, nil
}

func (r *CoursePermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.course.UserId, nil
}

func NewCourseRepo() *CourseRepo {
	return &CourseRepo{
		load: loader.NewCourseLoader(),
	}
}

type CourseRepo struct {
	load   *loader.CourseLoader
	permit *Permitter
}

func (r *CourseRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *CourseRepo) Close() {
	r.load.ClearAll()
}

func (r *CourseRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("course connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *CourseRepo) CountByApplee(
	ctx context.Context,
	appleeId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseByApplee(db, appleeId)
}

func (r *CourseRepo) CountByEnrollee(
	ctx context.Context,
	enrolleeId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseByEnrollee(db, enrolleeId)
}

func (r *CourseRepo) CountByTopic(
	ctx context.Context,
	topicId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseByTopic(db, topicId)
}

func (r *CourseRepo) CountBySearch(
	ctx context.Context,
	within *mytype.OID,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseBySearch(db, within, query)
}

func (r *CourseRepo) CountByStudy(
	ctx context.Context,
	studyId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseByStudy(db, studyId)
}

func (r *CourseRepo) CountByTopicSearch(
	ctx context.Context,
	query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseByTopicSearch(db, query)
}

func (r *CourseRepo) CountByUser(
	ctx context.Context,
	userId string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseByUser(db, userId)
}

func (r *CourseRepo) Create(
	ctx context.Context,
	c *data.Course,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, c); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(c.Name.String)
	innerSpace := regexp.MustCompile(`\s+`)
	if err := c.Name.Set(innerSpace.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	course, err := data.CreateCourse(db, c)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) Get(
	ctx context.Context,
	id string,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	course, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) GetByApplee(
	ctx context.Context,
	appleeId string,
	po *data.PageOptions,
) ([]*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courses, err := data.GetCourseByApplee(db, appleeId, po)
	if err != nil {
		return nil, err
	}
	coursePermits := make([]*CoursePermit, len(courses))
	if len(courses) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courses[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courses {
			coursePermits[i] = &CoursePermit{fieldPermFn, l}
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) GetByEnrollee(
	ctx context.Context,
	enrolleeId string,
	po *data.PageOptions,
) ([]*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courses, err := data.GetCourseByEnrollee(db, enrolleeId, po)
	if err != nil {
		return nil, err
	}
	coursePermits := make([]*CoursePermit, len(courses))
	if len(courses) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courses[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courses {
			coursePermits[i] = &CoursePermit{fieldPermFn, l}
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) GetByStudy(
	ctx context.Context,
	studyId string,
	po *data.PageOptions,
) ([]*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courses, err := data.GetCourseByStudy(db, studyId, po)
	if err != nil {
		return nil, err
	}
	coursePermits := make([]*CoursePermit, len(courses))
	if len(courses) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courses[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courses {
			coursePermits[i] = &CoursePermit{fieldPermFn, l}
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) GetByTopic(
	ctx context.Context,
	topicId string,
	po *data.PageOptions,
) ([]*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courses, err := data.GetCourseByTopic(db, topicId, po)
	if err != nil {
		return nil, err
	}
	coursePermits := make([]*CoursePermit, len(courses))
	if len(courses) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courses[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courses {
			coursePermits[i] = &CoursePermit{fieldPermFn, l}
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) GetByUser(
	ctx context.Context,
	userId string,
	po *data.PageOptions,
) ([]*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courses, err := data.GetCourseByUser(db, userId, po)
	if err != nil {
		return nil, err
	}
	coursePermits := make([]*CoursePermit, len(courses))
	if len(courses) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courses[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courses {
			coursePermits[i] = &CoursePermit{fieldPermFn, l}
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) GetByName(
	ctx context.Context,
	studyId,
	name string,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	course, err := r.load.GetByName(ctx, studyId, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) GetByNumber(
	ctx context.Context,
	studyId string,
	number int32,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	course, err := r.load.GetByNumber(ctx, studyId, number)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
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
		return nil, err
	}
	course, err := r.load.GetByStudyAndName(ctx, study, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
		return nil, err
	}
	return &CoursePermit{fieldPermFn, course}, nil
}

func (r *CourseRepo) Delete(
	ctx context.Context,
	course *data.Course,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, course); err != nil {
		return err
	}
	return data.DeleteCourse(db, course.Id.String)
}

func (r *CourseRepo) Search(
	ctx context.Context,
	within *mytype.OID,
	query string,
	po *data.PageOptions,
) ([]*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courses, err := data.SearchCourse(db, within, query, po)
	if err != nil {
		return nil, err
	}
	coursePermits := make([]*CoursePermit, len(courses))
	if len(courses) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courses[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courses {
			coursePermits[i] = &CoursePermit{fieldPermFn, l}
		}
	}
	return coursePermits, nil
}

func (r *CourseRepo) Update(
	ctx context.Context,
	c *data.Course,
) (*CoursePermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, c); err != nil {
		return nil, err
	}
	course, err := data.UpdateCourse(db, c)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, course)
	if err != nil {
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