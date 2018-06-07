package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewEVTLoader(
	svc *data.EVTService,
) *EVTLoader {
	return &EVTLoader{
		svc: svc,
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
						ks := splitCompositeKey(key)
						evt, err := svc.Get(ks[0], ks[1])
						results[i] = &dataloader.Result{Data: evt, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type EVTLoader struct {
	svc *data.EVTService

	batchGet *dataloader.Loader
}

func (r *EVTLoader) Clear(emailId, token string) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailId, token)
	r.batchGet.Clear(ctx, compositeKey)
}

func (r *EVTLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *EVTLoader) Get(emailId, token string) (*data.EVT, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailId, token)
	evtData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	evt, ok := evtData.(*data.EVT)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return evt, nil
}
