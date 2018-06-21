package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewEventLoader(svc *data.EventService) *EventLoader {
	return &EventLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetEventFn(svc.Get)),
	}
}

type EventLoader struct {
	svc *data.EventService

	batchGet *dataloader.Loader
}

func (r *EventLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *EventLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *EventLoader) Get(id string) (*data.Event, error) {
	ctx := context.Background()
	eventData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	event, ok := eventData.(*data.Event)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return event, nil
}

func (r *EventLoader) GetMany(ids *[]string) ([]*data.Event, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	eventData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	events := make([]*data.Event, len(eventData))
	for i, d := range eventData {
		var ok bool
		events[i], ok = d.(*data.Event)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return events, nil
}

func newBatchGetEventFn(
	getter func(string) (*data.Event, error),
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
				event, err := getter(key.String())
				results[i] = &dataloader.Result{Data: event, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
