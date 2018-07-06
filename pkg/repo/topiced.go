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

type TopicedPermit struct {
	checkFieldPermission FieldPermissionFunc
	topiced              *data.Topiced
}

func (r *TopicedPermit) Get() *data.Topiced {
	topiced := r.topiced
	fields := structs.Fields(topiced)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return topiced
}

func (r *TopicedPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.topiced.CreatedAt.Time, nil
}

func (r *TopicedPermit) ID() (n int32, err error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err = ErrAccessDenied
		return
	}
	n = r.topiced.Id.Int
	return
}

func (r *TopicedPermit) TopicId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("topic_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.topiced.TopicId, nil
}

func (r *TopicedPermit) TopicableId() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("topicable_id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.topiced.TopicableId, nil
}

func NewTopicedRepo(svc *data.TopicedService) *TopicedRepo {
	return &TopicedRepo{
		svc: svc,
	}
}

type TopicedRepo struct {
	load  *loader.TopicedLoader
	perms *Permitter
	svc   *data.TopicedService
}

func (r *TopicedRepo) Open(p *Permitter) error {
	r.perms = p
	if r.load == nil {
		r.load = loader.NewTopicedLoader(r.svc)
	}
	return nil
}

func (r *TopicedRepo) Close() {
	r.load.ClearAll()
}

func (r *TopicedRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("topiced connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *TopicedRepo) CountByTopic(
	topicId string,
) (int32, error) {
	return r.svc.CountByTopic(topicId)
}

func (r *TopicedRepo) CountByTopicable(
	topicableId string,
) (int32, error) {
	return r.svc.CountByTopicable(topicableId)
}

func (r *TopicedRepo) Connect(topiced *data.Topiced) (*TopicedPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(mytype.ConnectAccess, topiced); err != nil {
		return nil, err
	}
	topiced, err := r.svc.Connect(topiced)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, topiced)
	if err != nil {
		return nil, err
	}
	return &TopicedPermit{fieldPermFn, topiced}, nil
}

func (r *TopicedRepo) Get(t *data.Topiced) (*TopicedPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	var topiced *data.Topiced
	var err error
	if topiced.Id.Status != pgtype.Undefined {
		topiced, err = r.load.Get(t.Id.Int)
		if err != nil {
			return nil, err
		}
	} else if topiced.TopicableId.Status != pgtype.Undefined &&
		topiced.TopicId.Status != pgtype.Undefined {
		topiced, err = r.load.GetForTopicable(t.TopicableId.String, t.TopicId.String)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(
			"must include either topiced `id` or `topicable_id` and `topic_id` to get an topiced",
		)
	}
	fieldPermFn, err := r.perms.Check(mytype.ReadAccess, topiced)
	if err != nil {
		return nil, err
	}
	return &TopicedPermit{fieldPermFn, topiced}, nil
}

func (r *TopicedRepo) GetByTopic(
	topicId string,
	po *data.PageOptions,
) ([]*TopicedPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topiceds, err := r.svc.GetByTopic(topicId, po)
	if err != nil {
		return nil, err
	}
	topicedPermits := make([]*TopicedPermit, len(topiceds))
	if len(topiceds) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, topiceds[0])
		if err != nil {
			return nil, err
		}
		for i, t := range topiceds {
			topicedPermits[i] = &TopicedPermit{fieldPermFn, t}
		}
	}
	return topicedPermits, nil
}

func (r *TopicedRepo) GetByTopicable(
	topicableId string,
	po *data.PageOptions,
) ([]*TopicedPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topiceds, err := r.svc.GetByTopicable(topicableId, po)
	if err != nil {
		return nil, err
	}
	topicedPermits := make([]*TopicedPermit, len(topiceds))
	if len(topiceds) > 0 {
		fieldPermFn, err := r.perms.Check(mytype.ReadAccess, topiceds[0])
		if err != nil {
			return nil, err
		}
		for i, t := range topiceds {
			topicedPermits[i] = &TopicedPermit{fieldPermFn, t}
		}
	}
	return topicedPermits, nil
}

func (r *TopicedRepo) Disconnect(topiced *data.Topiced) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(mytype.DisconnectAccess, topiced); err != nil {
		return err
	}
	if topiced.Id.Status != pgtype.Undefined {
		return r.svc.Disconnect(topiced.Id.Int)
	} else if topiced.TopicableId.Status != pgtype.Undefined &&
		topiced.TopicId.Status != pgtype.Undefined {
		return r.svc.DisconnectFromTopicable(topiced.TopicableId.String, topiced.TopicId.String)
	}
	return errors.New(
		"must include either topiced `id` or `topicable_id` and `topic_id` to delete a topiced",
	)
}
