package repo

import (
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
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

func (r *LessonCommentPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.Id, nil
}

func (r *LessonCommentPermit) LessonId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("lesson_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.LessonId, nil
}

func (r *LessonCommentPermit) PublishedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("published_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.lessonComment.PublishedAt.Time, nil
}

func (r *LessonCommentPermit) StudyId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("study_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.StudyId, nil
}

func (r *LessonCommentPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.lessonComment.UserId, nil
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

func (r *LessonCommentRepo) CountByLesson(userId, studyId, lessonId string) (int32, error) {
	return r.svc.CountByLesson(userId, studyId, lessonId)
}

func (r *LessonCommentRepo) CountByStudy(userId, studyId string) (int32, error) {
	return r.svc.CountByStudy(userId, studyId)
}

func (r *LessonCommentRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *LessonCommentRepo) Create(lc *data.LessonComment) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(mytype.CreateAccess, lc); err != nil {
		return nil, err
	}
	lessonComment, err := r.svc.Create(lc)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) Get(id string) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessonComment, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) GetByLesson(
	userId,
	studyId,
	lessonId string,
	po *data.PageOptions,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessonComments, err := r.svc.GetByLesson(userId, studyId, lessonId, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, lessonComments[0])
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
	userId,
	studyId string,
	po *data.PageOptions,
) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessonComments, err := r.svc.GetByStudy(userId, studyId, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) GetByUser(userId string, po *data.PageOptions) ([]*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	lessonComments, err := r.svc.GetByUser(userId, po)
	if err != nil {
		return nil, err
	}
	lessonCommentPermits := make([]*LessonCommentPermit, len(lessonComments))
	if len(lessonComments) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, lessonComments[0])
		if err != nil {
			return nil, err
		}
		for i, l := range lessonComments {
			lessonCommentPermits[i] = &LessonCommentPermit{fieldPermFn, l}
		}
	}
	return lessonCommentPermits, nil
}

func (r *LessonCommentRepo) Delete(lessonComment *data.LessonComment) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.permit.Check(mytype.DeleteAccess, lessonComment); err != nil {
		return err
	}
	return r.svc.Delete(lessonComment.Id.String)
}

func (r *LessonCommentRepo) Update(lc *data.LessonComment) (*LessonCommentPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(mytype.UpdateAccess, lc); err != nil {
		return nil, err
	}
	lessonComment, err := r.svc.Update(lc)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, lessonComment)
	if err != nil {
		return nil, err
	}
	return &LessonCommentPermit{fieldPermFn, lessonComment}, nil
}

func (r *LessonCommentRepo) ViewerCanDelete(l *data.LessonComment) bool {
	if _, err := r.permit.Check(mytype.DeleteAccess, l); err != nil {
		return false
	}
	return true
}

func (r *LessonCommentRepo) ViewerCanUpdate(l *data.LessonComment) bool {
	if _, err := r.permit.Check(mytype.UpdateAccess, l); err != nil {
		return false
	}
	return true
}
