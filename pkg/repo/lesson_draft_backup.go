package repo

import (
	"context"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type LessonDraftBackupPermit struct {
	checkFieldPermission FieldPermissionFunc
	lesson               *data.LessonDraftBackup
}

func (r *LessonDraftBackupPermit) Get() *data.LessonDraftBackup {
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

func (r *LessonDraftBackupPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lesson.CreatedAt.Time, nil
}

func (r *LessonDraftBackupPermit) Draft() (string, error) {
	if ok := r.checkFieldPermission("draft"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.lesson.Draft.String, nil
}

func (r *LessonDraftBackupPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.lesson.ID.Int, nil
}

func (r *LessonDraftBackupPermit) LessonID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lesson.LessonID, nil
}

func (r *LessonDraftBackupPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lesson.UpdatedAt.Time, nil
}

func NewLessonDraftBackupRepo(conf *myconf.Config) *LessonDraftBackupRepo {
	return &LessonDraftBackupRepo{
		conf: conf,
		load: loader.NewLessonDraftBackupLoader(),
	}
}

type LessonDraftBackupRepo struct {
	conf   *myconf.Config
	load   *loader.LessonDraftBackupLoader
	permit *Permitter
}

func (r *LessonDraftBackupRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	lessons []*data.LessonDraftBackup,
) ([]*LessonDraftBackupPermit, error) {
	lessonPermits := make([]*LessonDraftBackupPermit, 0, len(lessons))
	for _, l := range lessons {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			lessonPermits = append(lessonPermits, &LessonDraftBackupPermit{fieldPermFn, l})
		}
	}
	return lessonPermits, nil
}

func (r *LessonDraftBackupRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *LessonDraftBackupRepo) Clear(id string) {
	r.load.Clear(id)
}

func (r *LessonDraftBackupRepo) Close() {
	r.load.ClearAll()
}

func (r *LessonDraftBackupRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *LessonDraftBackupRepo) Get(
	ctx context.Context,
	lessonID string,
	id int32,
) (*LessonDraftBackupPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, err := r.load.Get(ctx, lessonID, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonDraftBackupPermit{fieldPermFn, lesson}, nil
}

// Same as Get(), but doesn't use the dataloader
func (r *LessonDraftBackupRepo) Pull(
	ctx context.Context,
	lessonID string,
	id int32,
) (*LessonDraftBackupPermit, error) {
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
	lesson, err := data.GetLessonDraftBackup(db, lessonID, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lesson)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonDraftBackupPermit{fieldPermFn, lesson}, nil
}

func (r *LessonDraftBackupRepo) GetByLesson(
	ctx context.Context,
	lessonID string,
) ([]*LessonDraftBackupPermit, error) {
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
	lessons, err := data.GetLessonDraftBackupByLesson(db, lessonID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessons)
}

func (r *LessonDraftBackupRepo) Restore(
	ctx context.Context,
	lesson *data.Lesson,
	backupID int32,
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, lesson); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.RestoreLessonDraftFromBackup(db, lesson.ID.String, backupID)
}
