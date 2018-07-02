package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewLabelLoader(svc *data.LabelService) *LabelLoader {
	return &LabelLoader{
		svc:            svc,
		batchGet:       createLoader(newBatchGetLabelBy1Fn(svc.Get)),
		batchGetByName: createLoader(newBatchGetLabelBy1Fn(svc.GetByName)),
	}
}

type LabelLoader struct {
	svc *data.LabelService

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

func (r *LabelLoader) Get(id string) (*data.Label, error) {
	ctx := context.Background()
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

func (r *LabelLoader) GetByName(name string) (*data.Label, error) {
	ctx := context.Background()
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

func newBatchGetLabelBy1Fn(
	getter func(string) (*data.Label, error),
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
				label, err := getter(key.String())
				results[i] = &dataloader.Result{Data: label, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
