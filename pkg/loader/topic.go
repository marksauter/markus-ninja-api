package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewTopicLoader(svc *data.TopicService) *TopicLoader {
	return &TopicLoader{
		svc:            svc,
		batchGet:       createLoader(newBatchGetTopicBy1Fn(svc.Get)),
		batchGetByName: createLoader(newBatchGetTopicBy1Fn(svc.GetByName)),
	}
}

type TopicLoader struct {
	svc *data.TopicService

	batchGet       *dataloader.Loader
	batchGetByName *dataloader.Loader
}

func (r *TopicLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *TopicLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *TopicLoader) Get(id string) (*data.Topic, error) {
	ctx := context.Background()
	topicData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	topic, ok := topicData.(*data.Topic)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return topic, nil
}

func (r *TopicLoader) GetByName(name string) (*data.Topic, error) {
	ctx := context.Background()
	topicData, err := r.batchGetByName.Load(ctx, dataloader.StringKey(name))()
	if err != nil {
		return nil, err
	}
	topic, ok := topicData.(*data.Topic)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return topic, nil
}

func (r *TopicLoader) GetMany(ids *[]string) ([]*data.Topic, []error) {
	ctx := context.Background()
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

func newBatchGetTopicBy1Fn(
	getter func(string) (*data.Topic, error),
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
				topic, err := getter(key.String())
				results[i] = &dataloader.Result{Data: topic, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
