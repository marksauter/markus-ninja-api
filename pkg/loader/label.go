package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewLabelLoader() *LabelLoader {
	return &LabelLoader{
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
						label, err := data.GetLabel(db, key.String())
						results[i] = &dataloader.Result{Data: label, Error: err}
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
						label, err := data.GetLabelByName(db, key.String())
						results[i] = &dataloader.Result{Data: label, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type LabelLoader struct {
	batchGet       *dataloader.Loader
	batchGetByName *dataloader.Loader
}

func (r *LabelLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LabelLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByName.ClearAll()
}

func (r *LabelLoader) Get(
	ctx context.Context,
	id string,
) (*data.Label, error) {
	labelData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	label, ok := labelData.(*data.Label)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return label, nil
}

func (r *LabelLoader) GetByName(
	ctx context.Context,
	name string,
) (*data.Label, error) {
	labelData, err := r.batchGetByName.Load(ctx, dataloader.StringKey(name))()
	if err != nil {
		return nil, err
	}
	label, ok := labelData.(*data.Label)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return label, nil
}

func (r *LabelLoader) GetMany(ids *[]string) ([]*data.Label, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	labelData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	labels := make([]*data.Label, len(labelData))
	for i, d := range labelData {
		var ok bool
		labels[i], ok = d.(*data.Label)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return labels, nil
}
