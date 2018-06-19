package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewRefLoader(svc *data.RefService) *RefLoader {
	return &RefLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetRefFn(svc.Get)),
	}
}

type RefLoader struct {
	svc *data.RefService

	batchGet *dataloader.Loader
}

func (r *RefLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *RefLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *RefLoader) Get(id string) (*data.Ref, error) {
	ctx := context.Background()
	refData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	ref, ok := refData.(*data.Ref)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return ref, nil
}

func (r *RefLoader) GetMany(ids *[]string) ([]*data.Ref, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	refData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	refs := make([]*data.Ref, len(refData))
	for i, d := range refData {
		var ok bool
		refs[i], ok = d.(*data.Ref)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return refs, nil
}

func newBatchGetRefFn(
	getter func(string) (*data.Ref, error),
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
				ref, err := getter(key.String())
				results[i] = &dataloader.Result{Data: ref, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
