package repo

import (
	"errors"
	"time"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

type EnrolledPermit struct {
	checkFieldPermission FieldPermissionFunc
	enrolled             *data.Enrolled
}

func (r *EnrolledPermit) Get() *data.Enrolled {
	enrolled := r.enrolled
	fields := structs.Fields(enrolled)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return enrolled
}

func (r *EnrolledPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.enrolled.CreatedAt.Time, nil
}

func (r *EnrolledPermit) EnrollableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("enrollable_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.EnrollableId, nil
}

func (r *EnrolledPermit) ID() (n int32, err error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.enrolled.Id.Int
	return
}

func (r *EnrolledPermit) UserId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("user_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.enrolled.UserId, nil
}

func NewEnrolledRepo() *EnrolledRepo {
	return &EnrolledRepo{
		load: loader.NewEnrolledLoader(),
	}
}

type EnrolledRepo struct {
	load  *loader.EnrolledLoader
	perms *Permitter
}

func (r *EnrolledRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.perms = p
	return nil
}

func (r *EnrolledRepo) Close() {
	r.load.ClearAll()
}

func (r *EnrolledRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("enrolled connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *EnrolledRepo) CountByEnrollable(
	enrollableId string,
) (int32, error) {
	return data.CountEnrolledByEnrollable(enrollableId)
}

func (r *EnrolledRepo) CountByUser(
	userId string,
) (int32, error) {
	return data.CountEnrolledByUser(userId)
}

func (r *EnrolledRepo) Connect(enrolled *data.Enrolled) (*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(mytype.ConnectAccess, enrolled); err != nil {
		return nil, err
	}
	enrolled, err := r.svc.Connect(enrolled)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, enrolled)
	if err != nil {
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) Get(e *data.Enrolled) (*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var enrolled *data.Enrolled
	var err error
	if enrolled.Id.Status != pgtype.Undefined {
		enrolled, err = r.load.Get(e.Id.Int)
		if err != nil {
			return nil, err
		}
	} else if enrolled.EnrollableId.Status != pgtype.Undefined &&
		enrolled.UserId.Status != pgtype.Undefined {
		enrolled, err = r.load.GetForEnrollable(e.EnrollableId.String, e.UserId.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either enrolled `id` or `enrollable_id` and `user_id` to get an enrolled",
		)
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, enrolled)
	if err != nil {
		return nil, err
	}
	return &EnrolledPermit{fieldPermFn, enrolled}, nil
}

func (r *EnrolledRepo) GetByEnrollable(
	enrollableId string,
	po *data.PageOptions,
) ([]*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	enrolleds, err := r.svc.GetByEnrollable(enrollableId, po)
	if err != nil {
		return nil, err
	}
	enrolledPermits := make([]*EnrolledPermit, len(enrolleds))
	if len(enrolleds) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, enrolleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range enrolleds {
			enrolledPermits[i] = &EnrolledPermit{fieldPermFn, l}
		}
	}
	return enrolledPermits, nil
}

func (r *EnrolledRepo) GetByUser(
	userId string,
	po *data.PageOptions,
) ([]*EnrolledPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	enrolleds, err := r.svc.GetByUser(userId, po)
	if err != nil {
		return nil, err
	}
	enrolledPermits := make([]*EnrolledPermit, len(enrolleds))
	if len(enrolleds) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, enrolleds[0])
		if err != nil {
			return nil, err
		}
		for i, l := range enrolleds {
			enrolledPermits[i] = &EnrolledPermit{fieldPermFn, l}
		}
	}
	return enrolledPermits, nil
}

func (r *EnrolledRepo) Disconnect(enrolled *data.Enrolled) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(mytype.DisconnectAccess, enrolled); err != nil {
		return err
	}
	if enrolled.Id.Status != pgtype.Undefined {
		return r.svc.Disconnect(enrolled.Id.Int)
	} else if enrolled.EnrollableId.Status != pgtype.Undefined &&
		enrolled.UserId.Status != pgtype.Undefined {
		return r.svc.DisconnectFromEnrollable(enrolled.EnrollableId.String, enrolled.UserId.String)
	}
	return errors.New("must include either `id` or `enrollable_id` and `user_id` to delete an enrolled")
}
