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

type CommentDraftBackupPermit struct {
	checkFieldPermission FieldPermissionFunc
	comment              *data.CommentDraftBackup
}

func (r *CommentDraftBackupPermit) Get() *data.CommentDraftBackup {
	comment := r.comment
	fields := structs.Fields(comment)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return comment
}

func (r *CommentDraftBackupPermit) CommentID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("comment_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.comment.CommentID, nil
}

func (r *CommentDraftBackupPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.comment.CreatedAt.Time, nil
}

func (r *CommentDraftBackupPermit) Draft() (string, error) {
	if ok := r.checkFieldPermission("draft"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return "", err
	}
	return r.comment.Draft.String, nil
}

func (r *CommentDraftBackupPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.comment.ID.Int, nil
}

func (r *CommentDraftBackupPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.comment.UpdatedAt.Time, nil
}

func NewCommentDraftBackupRepo(conf *myconf.Config) *CommentDraftBackupRepo {
	return &CommentDraftBackupRepo{
		conf: conf,
		load: loader.NewCommentDraftBackupLoader(),
	}
}

type CommentDraftBackupRepo struct {
	conf   *myconf.Config
	load   *loader.CommentDraftBackupLoader
	permit *Permitter
}

func (r *CommentDraftBackupRepo) filterPermittable(
	ctx context.Context,
	accessLevel mytype.AccessLevel,
	comments []*data.CommentDraftBackup,
) ([]*CommentDraftBackupPermit, error) {
	commentPermits := make([]*CommentDraftBackupPermit, 0, len(comments))
	for _, l := range comments {
		fieldPermFn, err := r.permit.Check(ctx, accessLevel, l)
		if err != nil {
			if err != ErrAccessDenied {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
		} else {
			commentPermits = append(commentPermits, &CommentDraftBackupPermit{fieldPermFn, l})
		}
	}
	return commentPermits, nil
}

func (r *CommentDraftBackupRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *CommentDraftBackupRepo) Clear(id string) {
	r.load.Clear(id)
}

func (r *CommentDraftBackupRepo) Close() {
	r.load.ClearAll()
}

func (r *CommentDraftBackupRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *CommentDraftBackupRepo) Get(
	ctx context.Context,
	commentID string,
	id int32,
) (*CommentDraftBackupPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	comment, err := r.load.Get(ctx, commentID, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, comment)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CommentDraftBackupPermit{fieldPermFn, comment}, nil
}

// Same as Get(), but doesn't use the dataloader
func (r *CommentDraftBackupRepo) Pull(
	ctx context.Context,
	commentID string,
	id int32,
) (*CommentDraftBackupPermit, error) {
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
	comment, err := data.GetCommentDraftBackup(db, commentID, id)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, comment)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &CommentDraftBackupPermit{fieldPermFn, comment}, nil
}

func (r *CommentDraftBackupRepo) GetByComment(
	ctx context.Context,
	commentID string,
) ([]*CommentDraftBackupPermit, error) {
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
	comments, err := data.GetCommentDraftBackupByComment(db, commentID)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return r.filterPermittable(ctx, mytype.ReadAccess, comments)
}

func (r *CommentDraftBackupRepo) Restore(
	ctx context.Context,
	comment *data.Comment,
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
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, comment); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return data.RestoreCommentDraftFromBackup(db, comment.ID.String, backupID)
}
