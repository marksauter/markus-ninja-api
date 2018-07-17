package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewTopicLoader() *TopicLoader {
	return &TopicLoader{
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
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						topic, err := data.GetTopic(db, key.String())
						results[i] = &dataloader.Result{Data: topic, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByName: createLoader(
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
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						topic, err := data.GetTopicByName(db, key.String())
						results[i] = &dataloader.Result{Data: topic, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type TopicLoader struct {
	batchGet       *dataloader.Loader
	batchGetByName *dataloader.Loader
}

func (r *TopicLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *TopicLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByName.ClearAll()
}

func (r *TopicLoader) Get(
	ctx context.Context,
	id string,
) (*data.Topic, error) {
	topicData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	topic, ok := topicData.(*data.Topic)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGetByName.Prime(ctx, dataloader.StringKey(topic.Name.String), topic)

	return topic, nil
}

func (r *TopicLoader) GetByName(
	ctx context.Context,
	id string,
) (*data.Topic, error) {
	topicData, err := r.batchGetByName.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	topic, ok := topicData.(*data.Topic)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(topic.Id.String), topic)

	return topic, nil
}

func (r *TopicLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Topic, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	topicData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	topics := make([]*data.Topic, len(topicData))
	for i, d := range topicData {
		var ok bool
		topics[i], ok = d.(*data.Topic)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return topics, nil
}
