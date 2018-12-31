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

type ActivityPermit struct {
	checkFieldPermission FieldPermissionFunc
	activity             *data.Activity
}

func (r *ActivityPermit) Get() *data.Activity {
	activity := r.activity
	fields := structs.Fields(activity)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return activity
}

func (r *ActivityPermit) AdvancedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("advanced_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	if r.activity.AdvancedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.activity.AdvancedAt.Time, nil
}

func (r *ActivityPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.activity.CreatedAt.Time, nil
}

func (r *ActivityPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.activity.Description.String, nil
}

func (r *ActivityPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.activity.ID, nil
}

func (r *ActivityPermit) LessonID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.activity.LessonID, nil
}

func (r *ActivityPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.activity.Name.String, nil
}

func (r *ActivityPermit) Number() (int32, error) {
	if ok := r.checkFieldPermission("number"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.activity.Number.Int, nil
}

func (r *ActivityPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.activity.StudyID, nil
}

func (r *ActivityPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.activity.UpdatedAt.Time, nil
}

func (r *ActivityPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.activity.UserID, nil
}

func NewActivityRepo(conf *myconf.Config) *ActivityRepo {
	return &ActivityRepo{
		conf: conf,
		load: loader.NewActivityLoader(),
	}
}

type ActivityRepo struct {
	conf   *myconf.Config
	load   *loader.ActivityLoader
	permit *Permitter
}

func (r *ActivityRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	activities []*data.Activity,
) ([]*ActivityPermit, error) {
	activityPermits := make([]*ActivityPermit, 0, len(activities))
	for _, l := range activities {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			activityPermits = append(activityPermits, &ActivityPermit{fieldPermFn, l})
		}
	}
	return activityPermits, nil
}

func (r *ActivityRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *ActivityRepo) Close() {
	r.load.ClearAll()
}

func (r *ActivityRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *ActivityRepo) CountByLesson(
	ctx context.Context,
	lessonID string,
	filters *data.ActivityFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountActivityByLesson(db, lessonID, filters)
}

func (r *ActivityRepo) CountBySearch(
	ctx context.Context,
	filters *data.ActivityFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountActivityBySearch(db, filters)
}

func (r *ActivityRepo) CountByStudy(
	ctx context.Context,
	studyID string,
	filters *data.ActivityFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountActivityByStudy(db, studyID, filters)
}

func (r *ActivityRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.ActivityFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountActivityByUser(db, userID, filters)
}

func (r *ActivityRepo) Create(
	ctx context.Context,
	c *data.Activity,
) (*ActivityPermit, error) {
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
	activity, err := data.CreateActivity(db, c)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) Get(
	ctx context.Context,
	id string,
) (*ActivityPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, err := r.load.Get(ctx, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) Pull(
	ctx context.Context,
	id string,
) (*ActivityPermit, error) {
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
	activity, err := data.GetActivity(db, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) GetByLesson(
	ctx context.Context,
	lessonID string,
	po *data.PageOptions,
	filters *data.ActivityFilterOptions,
) ([]*ActivityPermit, error) {
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
	activities, err := data.GetActivityByLesson(db, lessonID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, activities)
}

func (r *ActivityRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
	filters *data.ActivityFilterOptions,
) ([]*ActivityPermit, error) {
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
	activities, err := data.GetActivityByStudy(db, studyID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, activities)
}

func (r *ActivityRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.ActivityFilterOptions,
) ([]*ActivityPermit, error) {
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
	activities, err := data.GetActivityByUser(db, userID, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, activities)
}

func (r *ActivityRepo) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*ActivityPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, err := r.load.GetByName(ctx, studyID, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) GetByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (*ActivityPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, err := r.load.GetByNumber(ctx, studyID, number)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) GetByStudyAndName(
	ctx context.Context,
	study,
	name string,
) (*ActivityPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, err := r.load.GetByStudyAndName(ctx, study, name)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) Delete(
	ctx context.Context,
	activity *data.Activity,
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
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, activity); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.DeleteActivity(db, activity.ID.String)
}

func (r *ActivityRepo) Search(
	ctx context.Context,
	po *data.PageOptions,
	filters *data.ActivityFilterOptions,
) ([]*ActivityPermit, error) {
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
	activities, err := data.SearchActivity(db, po, filters)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, activities)
}

func (r *ActivityRepo) Update(
	ctx context.Context,
	c *data.Activity,
) (*ActivityPermit, error) {
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
	activity, err := data.UpdateActivity(db, c)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, activity)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &ActivityPermit{fieldPermFn, activity}, nil
}

func (r *ActivityRepo) ViewerCanAdmin(
	ctx context.Context,
	c *data.Activity,
) (bool, error) {
	return r.permit.ViewerCanAdmin(ctx, c)
}
