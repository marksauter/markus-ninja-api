package loader

import (
	"context"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewEVTLoader() *EVTLoader {
	return &EVTLoader{
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
						ks := splitCompositeKey(key)
						evt, err := data.GetEVT(db, ks[0], ks[1])
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
	batchGet *dataloader.Loader
}

func (r *EVTLoader) Clear(emailID, token string) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailID, token)
	r.batchGet.Clear(ctx, compositeKey)
}

func (r *EVTLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *EVTLoader) Get(
	ctx context.Context,
	emailID,
	token string,
) (*data.EVT, error) {
	compositeKey := newCompositeKey(emailID, token)
	evtData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	evt, ok := evtData.(*data.EVT)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return evt, nil
}
