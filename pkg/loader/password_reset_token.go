package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewPRTLoader(
	svc *data.PRTService,
) *PRTLoader {
	return &PRTLoader{
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
						prt, err := svc.GetByPK(ks[0], ks[1])
						results[i] = &dataloader.Result{Data: prt, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type PRTLoader struct {
	svc *data.PRTService

	batchGet *dataloader.Loader
}

func (r *PRTLoader) Clear(emailId, token string) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailId, token)
	r.batchGet.Clear(ctx, compositeKey)
}

func (r *PRTLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *PRTLoader) Get(emailId, token string) (*data.PRT, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailId, token)
	prtData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	prt, ok := prtData.(*data.PRT)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return prt, nil
}
