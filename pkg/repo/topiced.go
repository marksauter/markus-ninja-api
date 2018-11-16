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
	"github.com/marksauter/markus-ninja-api/pkg/util"
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
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return time.Time{}, err
	}
	return r.topiced.CreatedAt.Time, nil
}

func (r *TopicedPermit) ID() (int32, error) {
	if ok := r.checkFieldPermission("id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		var n int32
		return n, err
	}
	return r.topiced.ID.Int, nil
}

func (r *TopicedPermit) TopicID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("topic_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.topiced.TopicID, nil
}

func (r *TopicedPermit) TopicableID() (*mytype.OID, error) {
	if ok := r.checkFieldPermission("topicable_id"); !ok {
		err := ErrAccessDenied
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &r.topiced.TopicableID, nil
}

func NewTopicedRepo(conf *myconf.Config) *TopicedRepo {
	return &TopicedRepo{
		conf: conf,
		load: loader.NewTopicedLoader(),
	}
}

type TopicedRepo struct {
	conf   *myconf.Config
	load   *loader.TopicedLoader
	permit *Permitter
}

func (r *TopicedRepo) Open(p *Permitter) error {
	if p == nil {
		err := ErrNilPermitter
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	r.permit = p
	return nil
}

func (r *TopicedRepo) Close() {
	r.load.ClearAll()
}

func (r *TopicedRepo) CheckConnection() error {
	if r.load == nil {
		err := ErrConnClosed
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	return nil
}

// Service methods

func (r *TopicedRepo) CountByTopic(
	ctx context.Context,
	topicID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountTopicedByTopic(db, topicID)
}

func (r *TopicedRepo) CountByTopicable(
	ctx context.Context,
	topicableID string,
) (int32, error) {
	var n int32
	db, ok := myctx.QueryerFromContext(ctx)
	if !ok {
		err := &myctx.ErrNotFound{"queryer"}
		mylog.Log.WithError(err).Error(util.Trace(""))
		return n, err
	}
	return data.CountTopicedByTopicable(db, topicableID)
}

func (r *TopicedRepo) Connect(
	ctx context.Context,
	topiced *data.Topiced,
) (*TopicedPermit, error) {
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
	if _, err := r.permit.Check(ctx, mytype.ConnectAccess, topiced); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	topiced, err := data.CreateTopiced(db, *topiced)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topiced)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &TopicedPermit{fieldPermFn, topiced}, nil
}

func (r *TopicedRepo) Get(
	ctx context.Context,
	t *data.Topiced,
) (*TopicedPermit, error) {
	if err := r.CheckConnection(); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	var topiced *data.Topiced
	var err error
	if t.ID.Status != pgtype.Undefined {
		topiced, err = r.load.Get(ctx, t.ID.Int)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else if t.TopicableID.Status != pgtype.Undefined &&
		t.TopicID.Status != pgtype.Undefined {
		topiced, err = r.load.GetByTopicableAndTopic(
			ctx,
			t.TopicableID.String,
			t.TopicID.String,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else {
		err := errors.New(
			"must include either topiced `id` or `topicable_id` and `topic_id` to get an topiced",
		)
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topiced)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &TopicedPermit{fieldPermFn, topiced}, nil
}

func (r *TopicedRepo) GetByTopic(
	ctx context.Context,
	topicID string,
	po *data.PageOptions,
) ([]*TopicedPermit, error) {
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
	topiceds, err := data.GetTopicedByTopic(db, topicID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	topicedPermits := make([]*TopicedPermit, len(topiceds))
	if len(topiceds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topiceds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, t := range topiceds {
			topicedPermits[i] = &TopicedPermit{fieldPermFn, t}
		}
	}
	return topicedPermits, nil
}

func (r *TopicedRepo) GetByTopicable(
	ctx context.Context,
	topicableID string,
	po *data.PageOptions,
) ([]*TopicedPermit, error) {
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
	topiceds, err := data.GetTopicedByTopicable(db, topicableID, po)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	topicedPermits := make([]*TopicedPermit, len(topiceds))
	if len(topiceds) > 0 {
		fieldPermFn, err := r.permit.Check(ctx, mytype.ReadAccess, topiceds[0])
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		for i, t := range topiceds {
			topicedPermits[i] = &TopicedPermit{fieldPermFn, t}
		}
	}
	return topicedPermits, nil
}

func (r *TopicedRepo) Disconnect(
	ctx context.Context,
	topiced *data.Topiced,
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
	if _, err := r.permit.Check(ctx, mytype.DisconnectAccess, topiced); err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return err
	}
	if topiced.ID.Status != pgtype.Undefined {
		return data.DeleteTopiced(db, topiced.ID.Int)
	} else if topiced.TopicableID.Status != pgtype.Undefined &&
		topiced.TopicID.Status != pgtype.Undefined {
		return data.DeleteTopicedByTopicableAndTopic(db, topiced.TopicableID.String, topiced.TopicID.String)
	}
	err := errors.New(
		"must include either topiced `id` or `topicable_id` and `topic_id` to delete a topiced",
	)
	mylog.Log.WithError(err).Error(util.Trace(""))
	return err
}
