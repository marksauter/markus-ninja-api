package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewEventLoader() *EventLoader {
	return &EventLoader{
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
						event, err := data.GetEvent(db, key.String())
						results[i] = &dataloader.Result{Data: event, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetBySourceActionTarget: createLoader(
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
						event, err := data.GetEventBySourceActionTarget(db, ks[0], ks[1], ks[2])
						results[i] = &dataloader.Result{Data: event, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type EventLoader struct {
	batchGet                     *dataloader.Loader
	batchGetBySourceActionTarget *dataloader.Loader
}

func (r *EventLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *EventLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetBySourceActionTarget.ClearAll()
}

func (r *EventLoader) Get(
	ctx context.Context,
	id string,
) (*data.Event, error) {
	eventData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	event, ok := eventData.(*data.Event)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(
		event.SourceId.String,
		event.Action.String,
		event.TargetId.String,
	)
	r.batchGetBySourceActionTarget.Prime(ctx, compositeKey, event)

	return event, nil
}

func (r *EventLoader) GetBySourceActionTarget(
	ctx context.Context,
	sourceId,
	action,
	targetId string,
) (*data.Event, error) {
	compositeKey := newCompositeKey(sourceId, action, targetId)
	eventData, err := r.batchGetBySourceActionTarget.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	event, ok := eventData.(*data.Event)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(event.Id.String), event)

	return event, nil
}

func (r *EventLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Event, []error) {
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
