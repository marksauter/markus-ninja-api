package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type CourseLessonPermit struct {
	checkFieldPermission FieldPermissionFunc
	courseLesson         *data.CourseLesson
}

func (r *CourseLessonPermit) Get() *data.CourseLesson {
	courseLesson := r.courseLesson
	fields := structs.Fields(courseLesson)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return courseLesson
}

func (r *CourseLessonPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.courseLesson.CreatedAt.Time, nil
}

func (r *CourseLessonPermit) CourseID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("course_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.courseLesson.CourseID, nil
}

func (r *CourseLessonPermit) LessonID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.courseLesson.LessonID, nil
}

func (r *CourseLessonPermit) Number() (n int32, err error) {
	if ok := r.checkFieldPermission("number"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.courseLesson.Number.Int
	return
}

func NewCourseLessonRepo(conf *myconf.Config) *CourseLessonRepo {
	return &CourseLessonRepo{
		conf: conf,
		load: loader.NewCourseLessonLoader(),
	}
}

type CourseLessonRepo struct {
	conf   *myconf.Config
	load   *loader.CourseLessonLoader
	permit *Permitter
}

func (r *CourseLessonRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *CourseLessonRepo) Close() {
	r.load.ClearAll()
}

func (r *CourseLessonRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("courseLesson connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *CourseLessonRepo) CountByCourse(
	ctx context.Context,
	courseID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountCourseLessonByCourse(db, courseID)
}

func (r *CourseLessonRepo) Connect(
	ctx context.Context,
	courseLesson *data.CourseLesson,
) (*CourseLessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, courseLesson); err != nil {
		return nil, err
	}
	courseLesson, err := data.CreateCourseLesson(db, *courseLesson)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courseLesson)
	if err != nil {
		return nil, err
	}
	return &CourseLessonPermit{fieldPermFn, courseLesson}, nil
}

func (r *CourseLessonRepo) Get(
	ctx context.Context,
	lessonID string,
) (*CourseLessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	courseLesson, err := r.load.Get(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courseLesson)
	if err != nil {
		return nil, err
	}
	return &CourseLessonPermit{fieldPermFn, courseLesson}, nil
}

func (r *CourseLessonRepo) GetByCourseAndNumber(
	ctx context.Context,
	courseID string,
	number int32,
) (*CourseLessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	courseLesson, err := r.load.GetByCourseAndNumber(ctx, courseID, number)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courseLesson)
	if err != nil {
		return nil, err
	}
	return &CourseLessonPermit{fieldPermFn, courseLesson}, nil
}

func (r *CourseLessonRepo) GetByCourse(
	ctx context.Context,
	courseID string,
	po *data.PageOptions,
) ([]*CourseLessonPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	courseLessons, err := data.GetCourseLessonByCourse(db, courseID, po)
	if err != nil {
		return nil, err
	}
	courseLessonPermits := make([]*CourseLessonPermit, len(courseLessons))
	if len(courseLessons) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, courseLessons[0])
		if err != nil {
			return nil, err
		}
		for i, l := range courseLessons {
			courseLessonPermits[i] = &CourseLessonPermit{fieldPermFn, l}
		}
	}
	return courseLessonPermits, nil
}

func (r *CourseLessonRepo) Disconnect(
	ctx context.Context,
	courseLesson *data.CourseLesson,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, courseLesson); err != nil {
		return err
	}
	return data.DeleteCourseLesson(db, courseLesson.LessonID.String)
}

func (r *CourseLessonRepo) Move(
	ctx context.Context,
	courseLesson *data.CourseLesson,
	afterLessonID string,
) (*data.CourseLesson, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, courseLesson); err != nil {
		return nil, err
	}
	return data.MoveCourseLesson(
		db,
		courseLesson.CourseID.String,
		courseLesson.LessonID.String,
		afterLessonID,
	)
}
