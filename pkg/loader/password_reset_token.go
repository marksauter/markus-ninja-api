package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewPRTLoader() *PRTLoader {
	return &PRTLoader{
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
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						prt, err := data.GetPRT(db, ks[0], ks[1])
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
	batchGet *dataloader.Loader
}

func (r *PRTLoader) Clear(emailID, token string) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailID, token)
	r.batchGet.Clear(ctx, compositeKey)
}

func (r *PRTLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *PRTLoader) Get(
	ctx context.Context,
	emailID,
	token string,
) (*data.PRT, error) {
	compositeKey := newCompositeKey(emailID, token)
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
