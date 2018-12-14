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

type LessonCommentDraftBackupPermit struct {
	checkFieldPermission FieldPermissionFunc
	lessonComment        *data.LessonCommentDraftBackup
}

func (r *LessonCommentDraftBackupPermit) Get() *data.LessonCommentDraftBackup {
	lessonComment := r.lessonComment
	fields := structs.Fields(lessonComment)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return lessonComment
}

func (r *LessonCommentDraftBackupPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lessonComment.CreatedAt.Time, nil
}

func (r *LessonCommentDraftBackupPermit) Draft() (string, error) {
	if ok := r.checkFieldPermission("draft"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.lessonComment.Draft.String, nil
}

func (r *LessonCommentDraftBackupPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.lessonComment.ID.Int, nil
}

func (r *LessonCommentDraftBackupPermit) LessonCommentID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_comment_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.lessonComment.LessonCommentID, nil
}

func (r *LessonCommentDraftBackupPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.lessonComment.UpdatedAt.Time, nil
}

func NewLessonCommentDraftBackupRepo(conf *myconf.Config) *LessonCommentDraftBackupRepo {
	return &LessonCommentDraftBackupRepo{
		conf: conf,
		load: loader.NewLessonCommentDraftBackupLoader(),
	}
}

type LessonCommentDraftBackupRepo struct {
	conf   *myconf.Config
	load   *loader.LessonCommentDraftBackupLoader
	permit *Permitter
}

func (r *LessonCommentDraftBackupRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	lessonComments []*data.LessonCommentDraftBackup,
) ([]*LessonCommentDraftBackupPermit, error) {
	lessonCommentPermits := make([]*LessonCommentDraftBackupPermit, 0, len(lessonComments))
	for _, l := range lessonComments {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			lessonCommentPermits = append(lessonCommentPermits, &LessonCommentDraftBackupPermit{fieldPermFn, l})
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentDraftBackupRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *LessonCommentDraftBackupRepo) Clear(id string) {
	r.load.Clear(id)
}

func (r *LessonCommentDraftBackupRepo) Close() {
	r.load.ClearAll()
}

func (r *LessonCommentDraftBackupRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *LessonCommentDraftBackupRepo) Get(
	ctx context.Context,
	lessonCommentID string,
	id int32,
) (*LessonCommentDraftBackupPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessonComment, err := r.load.Get(ctx, lessonCommentID, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComment)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonCommentDraftBackupPermit{fieldPermFn, lessonComment}, nil
}

// Same as Get(), but doesn't use the dataloader
func (r *LessonCommentDraftBackupRepo) Pull(
	ctx context.Context,
	lessonCommentID string,
	id int32,
) (*LessonCommentDraftBackupPermit, error) {
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
	lessonComment, err := data.GetLessonCommentDraftBackup(db, lessonCommentID, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComment)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &LessonCommentDraftBackupPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentDraftBackupRepo) GetByLessonComment(
	ctx context.Context,
	lessonCommentID string,
) ([]*LessonCommentDraftBackupPermit, error) {
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
	lessonComments, err := data.GetLessonCommentDraftBackupByLessonComment(db, lessonCommentID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, lessonComments)
}

func (r *LessonCommentDraftBackupRepo) Restore(
	ctx context.Context,
	lessonComment *data.LessonComment,
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, lessonComment); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.RestoreLessonCommentDraftFromBackup(db, lessonComment.ID.String, backupID)
}
