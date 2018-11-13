package repo

import (
	"context"
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type StudyPermit struct {
	checkFieldPermission FieldPermissionFunc
	study                *data.Study
}

func (r *StudyPermit) Get() *data.Study {
	study := r.study
	fields := structs.Fields(study)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return study
}

func (r *StudyPermit) AdvancedAt() (*time.Time, error) {
	if ok := r.checkFieldPermission("advanced_at"); !ok {
		return nil, ErrAccessDenied
	}
	if r.study.AdvancedAt.Status == pgtype.Null {
		return nil, nil
	}
	return &r.study.AdvancedAt.Time, nil
}

func (r *StudyPermit) AppledAt() time.Time {
	return r.study.AppledAt.Time
}

func (r *StudyPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.study.CreatedAt.Time, nil
}

func (r *StudyPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Description.String, nil
}

func (r *StudyPermit) EnrolledAt() time.Time {
	return r.study.EnrolledAt.Time
}

func (r *StudyPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.study.ID, nil
}

func (r *StudyPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Name.String, nil
}

func (r *StudyPermit) Private() (bool, error) {
	if ok := r.checkFieldPermission("private"); !ok {
		return false, ErrAccessDenied
	}
	return r.study.Private.Bool, nil
}

func (r *StudyPermit) TopicedAt() time.Time {
	return r.study.TopicedAt.Time
}

func (r *StudyPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.study.UpdatedAt.Time, nil
}

func (r *StudyPermit) UserID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.study.UserID, nil
}

func NewStudyRepo(conf *myconf.Config) *StudyRepo {
	return &StudyRepo{
		conf: conf,
		load: loader.NewStudyLoader(),
	}
}

type StudyRepo struct {
	conf   *myconf.Config
	load   *loader.StudyLoader
	permit *Permitter
}

func (r *StudyRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *StudyRepo) Close() {
	r.load.ClearAll()
}

func (r *StudyRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("study connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *StudyRepo) CountByApplee(
	ctx context.Context,
	appleeID string,
	filters *data.StudyFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountStudyByApplee(db, appleeID, filters)
}

func (r *StudyRepo) CountByEnrollee(
	ctx context.Context,
	enrolleeID string,
	filters *data.StudyFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountStudyByEnrollee(db, enrolleeID, filters)
}

func (r *StudyRepo) CountByTopic(
	ctx context.Context,
	topicID string,
	filters *data.StudyFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountStudyByTopic(db, topicID, filters)
}

func (r *StudyRepo) CountBySearch(
	ctx context.Context,
	filters *data.StudyFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountStudyBySearch(db, filters)
}

func (r *StudyRepo) CountByUser(
	ctx context.Context,
	userID string,
	filters *data.StudyFilterOptions,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountStudyByUser(db, userID, filters)
}

func (r *StudyRepo) Create(
	ctx context.Context,
	s *data.Study,
) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.CreateAccess, s); err != nil {
		return nil, err
	}
	study, err := data.CreateStudy(db, s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Get(
	ctx context.Context,
	id string,
) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByApplee(
	ctx context.Context,
	appleeID string,
	po *data.PageOptions,
	filters *data.StudyFilterOptions,
) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	studies, err := data.GetStudyByApplee(db, appleeID, po, filters)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByEnrollee(
	ctx context.Context,
	enrolleeID string,
	po *data.PageOptions,
	filters *data.StudyFilterOptions,
) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	studies, err := data.GetStudyByEnrollee(db, enrolleeID, po, filters)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByTopic(
	ctx context.Context,
	topicID string,
	po *data.PageOptions,
	filters *data.StudyFilterOptions,
) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	studies, err := data.GetStudyByTopic(db, topicID, po, filters)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByUser(
	ctx context.Context,
	userID string,
	po *data.PageOptions,
	filters *data.StudyFilterOptions,
) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	studies, err := data.GetStudyByUser(db, userID, po, filters)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByName(
	ctx context.Context,
	userID,
	name string,
) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.GetByName(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByUserAndName(
	ctx context.Context,
	owner,
	name string,
) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.GetByUserAndName(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Delete(
	ctx context.Context,
	study *data.Study,
) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.DeleteAccess, study); err != nil {
		return err
	}
	return data.DeleteStudy(db, study.ID.String)
}

func (r *StudyRepo) Search(
	ctx context.Context,
	po *data.PageOptions,
	filters *data.StudyFilterOptions,
) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	studies, err := data.SearchStudy(db, po, filters)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) Update(
	ctx context.Context,
	s *data.Study,
) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	if _, err := r.permit.Check(ctx, mytype.UpdateAccess, s); err != nil {
		return nil, err
	}
	study, err := data.UpdateStudy(db, s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) ViewerCanAdmin(
	ctx context.Context,
	s *data.Study,
) (bool, error) {
	return r.permit.ViewerCanAdmin(ctx, s)
}
