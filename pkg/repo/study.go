package repo

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
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
	return &r.study.Id, nil
}

func (r *StudyPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.study.Name.String, nil
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

func (r *StudyPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.study.UserId, nil
}

func NewStudyRepo() *StudyRepo {
	return &StudyRepo{
		load: loader.NewStudyLoader(),
	}
}

type StudyRepo struct {
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

func (r *StudyRepo) CountByApplee(appleeId string) (int32, error) {
	return r.svc.CountByApplee(appleeId)
}

func (r *StudyRepo) CountByEnrollee(enrolleeId string) (int32, error) {
	return r.svc.CountByEnrollee(enrolleeId)
}

func (r *StudyRepo) CountByTopic(topicId string) (int32, error) {
	return r.svc.CountByTopic(topicId)
}

func (r *StudyRepo) CountBySearch(within *mytype.OID, query string) (int32, error) {
	return r.svc.CountBySearch(within, query)
}

func (r *StudyRepo) CountByUser(userId string) (int32, error) {
	return r.svc.CountByUser(userId)
}

func (r *StudyRepo) Create(s *data.Study) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(mytype.CreateAccess, s); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(s.Name.String)
	innerSpace := regexp.MustCompile(`\s+`)
	if err := s.Name.Set(innerSpace.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	study, err := r.svc.Create(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Get(id string) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByApplee(appleeId string, po *data.PageOptions) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByApplee(appleeId, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByEnrollee(enrolleeId string, po *data.PageOptions) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByEnrollee(enrolleeId, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByTopic(topicId string, po *data.PageOptions) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByTopic(topicId, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByUser(userId string, po *data.PageOptions) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.GetByUser(userId, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) GetByName(userId, name string) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.svc.GetByName(userId, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) GetByUserAndName(owner string, name string) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	study, err := r.load.GetByUserAndName(owner, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) Delete(study *data.Study) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.permit.Check(mytype.DeleteAccess, study); err != nil {
		return err
	}
	return r.svc.Delete(study.Id.String)
}

func (r *StudyRepo) Search(within *mytype.OID, query string, po *data.PageOptions) ([]*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	studies, err := r.svc.Search(within, query, po)
	if err != nil {
		return nil, err
	}
	studyPermits := make([]*StudyPermit, len(studies))
	if len(studies) > 0 {
		fieldPermFn, err := r.permit.Check(mytype.ReadAccess, studies[0])
		if err != nil {
			return nil, err
		}
		for i, l := range studies {
			studyPermits[i] = &StudyPermit{fieldPermFn, l}
		}
	}
	return studyPermits, nil
}

func (r *StudyRepo) Update(s *data.Study) (*StudyPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.permit.Check(mytype.UpdateAccess, s); err != nil {
		return nil, err
	}
	study, err := r.svc.Update(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(mytype.ReadAccess, study)
	if err != nil {
		return nil, err
	}
	return &StudyPermit{fieldPermFn, study}, nil
}

func (r *StudyRepo) ViewerCanDelete(s *data.Study) bool {
	if _, err := r.permit.Check(mytype.DeleteAccess, s); err != nil {
		return false
	}
	return true
}

func (r *StudyRepo) ViewerCanUpdate(s *data.Study) bool {
	if _, err := r.permit.Check(mytype.UpdateAccess, s); err != nil {
		return false
	}
	return true
}
