package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewLabeledLoader(svc *data.LabeledService) *LabeledLoader {
	return &LabeledLoader{
		svc:                  svc,
		batchGet:             createLoader(newBatchGetLabeledBy1Fn(svc.Get)),
		batchGetForLabelable: createLoader(newBatchGetLabeledBy2Fn(svc.GetForLabelable)),
	}
}

type LabeledLoader struct {
	svc *data.LabeledService

	batchGet             *dataloader.Loader
	batchGetForLabelable *dataloader.Loader
}

func (r *LabeledLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LabeledLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *LabeledLoader) Get(id int32) (*data.Labeled, error) {
	ctx := context.Background()
	key := strconv.Itoa(int(id))
	labeledData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	labeled, ok := labeledData.(*data.Labeled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(labeled.LabelableId.String, labeled.LabelId.String)
	r.batchGetForLabelable.Prime(ctx, compositeKey, labeled)

	return labeled, nil
}

func (r *LabeledLoader) GetForLabelable(labelableId, userId string) (*data.Labeled, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(labelableId, userId)
	labeledData, err := r.batchGetForLabelable.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	labeled, ok := labeledData.(*data.Labeled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(labeled.Id.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), labeled)

	return labeled, nil
}

func newBatchGetLabeledBy1Fn(
	getter func(int32) (*data.Labeled, error),
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
				id, err := strconv.ParseInt(key.String(), 10, 32)
				if err != nil {
					results[i] = &dataloader.Result{Error: err}
					return
				}
				labeled, err := getter(int32(id))
				results[i] = &dataloader.Result{Data: labeled, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}

func newBatchGetLabeledBy2Fn(
	getter func(string, string) (*data.Labeled, error),
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
				ks := splitCompositeKey(key)
				labeled, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: labeled, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
