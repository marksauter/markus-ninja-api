package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type LessonCommentPermit struct {
	checkFieldPermission FieldPermissionFunc
	lessonComment        *data.LessonComment
}

func (r *LessonCommentPermit) Get() *data.LessonComment {
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

func (r *LessonCommentPermit) Body() (*mytype.Markdown, error) {
	if ok := r.checkFieldPermission("body"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.Body, nil
}

func (r *LessonCommentPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.CreatedAt.Time, nil
}

func (r *LessonCommentPermit) Draft() (string, error) {
	if ok := r.checkFieldPermission("draft"); !ok {
		return "", ErrAccessDenied
	}
	return r.lessonComment.Draft.String, nil
}

func (r *LessonCommentPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.ID, nil
}

func (r *LessonCommentPermit) LastEditedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("last_edited_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.LastEditedAt.Time, nil
}

func (r *LessonCommentPermit) LessonID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.LessonID, nil
}

func (r *LessonCommentPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.PublishedAt.Time, nil
}

func (r *LessonCommentPermit) StudyID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.StudyID, nil
}

func (r *LessonCommentPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.UserID, nil
}

func (r *LessonCommentPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.UpdatedAt.Time, nil
}

func NewLessonCommentRepo() *LessonCommentRepo {
	return &LessonCommentRepo{
		load: loader.NewLessonCommentLoader(),
	}
}

type LessonCommentRepo struct {
	load   *loader.LessonCommentLoader
	permit *Permitter
}

func (r *LessonCommentRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *LessonCommentRepo) Close() {
	r.load.ClearAll()
}

func (r *LessonCommentRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("lesson_comment connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *LessonCommentRepo) CountByLesson(
	ctx context.Context,
	lessonID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonCommentByLesson(db, lessonID)
}

func (r *LessonCommentRepo) CountByStudy(
	ctx context.Context,
	studyID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonCommentByStudy(db, studyID)
}

func (r *LessonCommentRepo) CountByUser(
	ctx context.Context,
	userID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountLessonCommentByUser(db, userID)
}

func (r *LessonCommentRepo) Create(
	ctx context.Context,
	lc *data.LessonComment,
) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, lc); err != nil {
		return nil, err
	}
	lessonComment, err := data.CreateLessonComment(db, lc)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) Get(
	ctx context.Context,
	id string,
) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessonComment, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) BatchGet(
	ctx context.Context,
	ids []string,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessonComments, err := data.BatchGetLessonComment(db, ids)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetUserNewComment(
	ctx context.Context,
	userID,
	lessonID string,
) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessonComment, err := data.GetUserNewLessonComment(db, userID, lessonID)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) GetByLabel(
	ctx context.Context,
	labelID string,
	po *data.PageOptions,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessonComments, err := data.GetLessonCommentByLabel(db, labelID, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetByLesson(
	ctx context.Context,
	lessonID string,
	po *data.PageOptions,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessonComments, err := data.GetLessonCommentByLesson(db, lessonID, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetByStudy(
	ctx context.Context,
	studyID string,
	po *data.PageOptions,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessonComments, err := data.GetLessonCommentByStudy(db, studyID, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	lessonComments, err := data.GetLessonCommentByUser(db, userID, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) Delete(
	ctx context.Context,
	lc *data.LessonComment,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, lc); err != nil {
		return err
	}
	return data.DeleteLessonComment(db, lc.ID.String)
}

func (r *LessonCommentRepo) Update(
	ctx context.Context,
	lc *data.LessonComment,
) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, lc); err != nil {
		return nil, err
	}
	lessonComment, err := data.UpdateLessonComment(db, lc)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) ViewerCanDelete(
	ctx context.Context,
	l *data.LessonComment,
) bool {
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *LessonCommentRepo) ViewerCanUpdate(
	ctx context.Context,
	l *data.LessonComment,
) bool {
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
