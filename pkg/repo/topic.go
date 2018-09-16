package repo

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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
	return &r.topic.ID, nil
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

func NewTopicRepo() *TopicRepo {
	return &TopicRepo{
		load: loader.NewTopicLoader(),
	}
}

type TopicRepo struct {
	load   *loader.TopicLoader
	permit *Permitter
}

func (r *TopicRepo) Open(p *Permitter) error {
	if p == nil {
		return errors.New("permitter must not be nil")
	}
	r.permit = p
	return nil
}

func (r *TopicRepo) Close() {
	r.load.ClearAll()
}

func (r *TopicRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("topic connection closed")
		return ErrConnClosed
	}
	return nil
}

// Service methods

func (r *TopicRepo) CountBySearch(
	ctx context.Context,
	within *mytype.OID, query string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountTopicBySearch(db, within, query)
}

func (r *TopicRepo) CountByTopicable(
	ctx context.Context,
	topicableID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return n, &myctx.ErrNotFound{"queryer"}
	}
	return data.CountTopicByTopicable(db, topicableID)
}

func (r *TopicRepo) Create(
	ctx context.Context,
	s *data.Topic,
) (*TopicPermit, error) {
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
	name := strings.TrimSpace(s.Name.String)
	innerSpace := regexp.MustCompile(`\s+`)
	if err := s.Name.Set(innerSpace.ReplaceAllString(name, "-")); err != nil {
		return nil, err
	}
	topic, err := data.CreateTopic(db, s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

func (r *TopicRepo) Get(
	ctx context.Context,
	id string,
) (*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topic, err := r.load.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

func (r *TopicRepo) GetByTopicable(
	ctx context.Context,
	topicableID string,
	po *data.PageOptions,
) ([]*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	topics, err := data.GetTopicByTopicable(db, topicableID, po)
	if err != nil {
		return nil, err
	}
	topicPermits := make([]*TopicPermit, len(topics))
	if len(topics) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topics[0])
		if err != nil {
			return nil, err
		}
		for i, l := range topics {
			topicPermits[i] = &TopicPermit{fieldPermFn, l}
		}
	}
	return topicPermits, nil
}

func (r *TopicRepo) GetByName(
	ctx context.Context,
	name string,
) (*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	topic, err := r.load.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}

func (r *TopicRepo) Search(
	ctx context.Context,
	query string,
	po *data.PageOptions,
) ([]*TopicPermit, error) {
	if err := r.CheckConnection(); err != nil {
		return nil, err
	}
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"queryer"}
	}
	topics, err := data.SearchTopic(db, query, po)
	if err != nil {
		return nil, err
	}
	topicPermits := make([]*TopicPermit, len(topics))
	if len(topics) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topics[0])
		if err != nil {
			return nil, err
		}
		for i, l := range topics {
			topicPermits[i] = &TopicPermit{fieldPermFn, l}
		}
	}
	return topicPermits, nil
}

func (r *TopicRepo) Update(
	ctx context.Context,
	s *data.Topic,
) (*TopicPermit, error) {
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
	topic, err := data.UpdateTopic(db, s)
	if err != nil {
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topic)
	if err != nil {
		return nil, err
	}
	return &TopicPermit{fieldPermFn, topic}, nil
}
