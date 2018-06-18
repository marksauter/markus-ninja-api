package repo

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

type TopicPermit struct {
	checkFieldPermission FieldPermissionFunc
	topic                *data.Topic
}

func (r *TopicPermit) Get() *data.Topic {
	topic := r.topic
	fields := structs.Fields(topic)
	for _, f := range fields {
		name := f.Tag("db")
		if ok := r.checkFieldPermission(name); !ok {
			f.Zero()
		}
	}
	return topic
}

func (r *TopicPermit) CreatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("created_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.topic.CreatedAt.Time, nil
}

func (r *TopicPermit) Description() (string, error) {
	if ok := r.checkFieldPermission("description"); !ok {
		return "", ErrAccessDenied
	}
	return r.topic.Description.String, nil
}

func (r *TopicPermit) ID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		return nil, ErrAccessDenied
	}
	return &r.topic.Id, nil
}

func (r *TopicPermit) Name() (string, error) {
	if ok := r.checkFieldPermission("name"); !ok {
		return "", ErrAccessDenied
	}
	return r.topic.Name.String, nil
}

func (r *TopicPermit) UpdatedAt() (time.Time, error) {
	if ok := r.checkFieldPermission("updated_at"); !ok {
		return time.Time{}, ErrAccessDenied
	}
	return r.topic.UpdatedAt.Time, nil
}

func NewTopicRepo(perms *PermRepo, svc *data.TopicService) *TopicRepo {
	return &TopicRepo{
		perms: perms,
		svc:   svc,
	}
}

type TopicRepo struct {
	load  *loader.TopicLoader
	perms *PermRepo
	svc   *data.TopicService
}

func (r *TopicRepo) Open(ctx context.Context) error {
	err := r.perms.Open(ctx)
	if err != nil {
		return err
	}
	if r.load == nil {
		r.load = loader.NewTopicLoader(r.svc)
	}
	return nil
}

func (r *TopicRepo) Close() {
	r.load = nil
}

func (r *TopicRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("topic connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *TopicRepo) CountBySearch(within *mytype.OID, query string) (int32, error) {
	return r.svc.CountBySearch(within, query)
}

func (r *TopicRepo) CountByStudy(studyId string) (int32, error) {
	return r.svc.CountByStudy(studyId)
}

func (r *TopicRepo) Create(s *data.Topic) (*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Create, s); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(s.Name.String)
	innerSpace := regexp.MustCompile(`\s+`)
	if err := s.Name.Set(innerSpace.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	topic, err := r.svc.Create(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

func (r *TopicRepo) Get(id string) (*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topic, err := r.load.Get(id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

func (r *TopicRepo) GetByStudy(studyId string, po *data.PageOptions) ([]*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topics, err := r.svc.GetByStudy(studyId, po)
	if err != nil {
		return nil, err
	}
	topicPermits := make([]*TopicPermit, len(topics))
	if len(topics) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, topics[0])
		if err != nil {
			return nil, err
		}
		for i, l := range topics {
			topicPermits[i] = &TopicPermit{fieldPermFn, l}
		}
	}
	return topicPermits, nil
}

func (r *TopicRepo) GetByName(name string) (*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topic, err := r.load.GetByName(name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

func (r *TopicRepo) DeleteStudyRelation(topic *data.Topic) error {
	if err := r.CheckConnection(); err != nil {
		return err
	}
	if _, err := r.perms.Check(perm.Delete, topic); err != nil {
		return err
	}
	return r.svc.DeleteStudyRelation(topic.StudyId.String, topic.Id.String)
}

func (r *TopicRepo) Search(query string, po *data.PageOptions) ([]*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topics, err := r.svc.Search(query, po)
	if err != nil {
		return nil, err
	}
	topicPermits := make([]*TopicPermit, len(topics))
	if len(topics) > 0 {
		fieldPermFn, err := r.perms.Check(perm.Read, topics[0])
		if err != nil {
			return nil, err
		}
		for i, l := range topics {
			topicPermits[i] = &TopicPermit{fieldPermFn, l}
		}
	}
	return topicPermits, nil
}

func (r *TopicRepo) Update(s *data.Topic) (*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	if _, err := r.perms.Check(perm.Update, s); err != nil {
		return nil, err
	}
	topic, err := r.svc.Update(s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.perms.Check(perm.Read, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

// Middleware
func (r *TopicRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
