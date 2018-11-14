package loader

import (
	"context"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewTopicedLoader() *TopicedLoader {
	return &TopicedLoader{
		batchGet: createLoader(
			func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
				var (
					n       = len(keys)
					results = make([]*dataloader.Result, n)
					wg      sync.WaitGroup
				)

				wg.Add(n)

				for i, key := range keys {
					go func(i int, key dataloader.Key) {
						defer wg.Done()
						id, err := strconv.ParseInt(key.String(), 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						topiced, err := data.GetTopiced(db, int32(id))
						results[i] = &dataloader.Result{Data: topiced, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByTopicableAndTopic: createLoader(
			func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
				var (
					n       = len(keys)
					results = make([]*dataloader.Result, n)
					wg      sync.WaitGroup
				)

				wg.Add(n)

				for i, key := range keys {
					go func(i int, key dataloader.Key) {
						defer wg.Done()
						ks := splitCompositeKey(key)
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						topiced, err := data.GetTopicedByTopicableAndTopic(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: topiced, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type TopicedLoader struct {
	batchGet                    *dataloader.Loader
	batchGetByTopicableAndTopic *dataloader.Loader
}

func (r *TopicedLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *TopicedLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByTopicableAndTopic.ClearAll()
}

func (r *TopicedLoader) Get(
	ctx context.Context,
	id int32,
) (*data.Topiced, error) {
	key := strconv.Itoa(int(id))
	topicedData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	topiced, ok := topicedData.(*data.Topiced)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	compositeKey := newCompositeKey(topiced.TopicableID.String, topiced.TopicID.String)
	r.batchGetByTopicableAndTopic.Prime(ctx, compositeKey, topiced)

	return topiced, nil
}

func (r *TopicedLoader) GetByTopicableAndTopic(
	ctx context.Context,
	topicableID,
	userID string,
) (*data.Topiced, error) {
	compositeKey := newCompositeKey(topicableID, userID)
	topicedData, err := r.batchGetByTopicableAndTopic.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	topiced, ok := topicedData.(*data.Topiced)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	key := strconv.Itoa(int(topiced.ID.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), topiced)

	return topiced, nil
}
