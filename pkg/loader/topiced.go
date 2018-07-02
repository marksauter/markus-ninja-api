package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewTopicedLoader(svc *data.TopicedService) *TopicedLoader {
	return &TopicedLoader{
		svc:                  svc,
		batchGet:             createLoader(newBatchGetTopicedBy1Fn(svc.Get)),
		batchGetForTopicable: createLoader(newBatchGetTopicedBy2Fn(svc.GetForTopicable)),
	}
}

type TopicedLoader struct {
	svc *data.TopicedService

	batchGet             *dataloader.Loader
	batchGetForTopicable *dataloader.Loader
}

func (r *TopicedLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *TopicedLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetForTopicable.ClearAll()
}

func (r *TopicedLoader) Get(id int32) (*data.Topiced, error) {
	ctx := context.Background()
	key := strconv.Itoa(int(id))
	topicedData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	topiced, ok := topicedData.(*data.Topiced)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(topiced.TopicableId.String, topiced.TopicId.String)
	r.batchGetForTopicable.Prime(ctx, compositeKey, topiced)

	return topiced, nil
}

func (r *TopicedLoader) GetForTopicable(topicableId, topicId string) (*data.Topiced, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(topicableId, topicId)
	topicedData, err := r.batchGetForTopicable.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	topiced, ok := topicedData.(*data.Topiced)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(topiced.Id.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), topiced)

	return topiced, nil
}

func newBatchGetTopicedBy1Fn(
	getter func(int32) (*data.Topiced, error),
) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
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
				topiced, err := getter(int32(id))
				results[i] = &dataloader.Result{Data: topiced, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}

func newBatchGetTopicedBy2Fn(
	getter func(string, string) (*data.Topiced, error),
) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
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
				topiced, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: topiced, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
