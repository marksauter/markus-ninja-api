package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewLabeledLoader() *LabeledLoader {
	return &LabeledLoader{
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
						id, err := strconv.ParseInt(key.String(), 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						labeled, err := data.GetLabeled(db, int32(id))
						results[i] = &dataloader.Result{Data: labeled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByLabelableAndLabel: createLoader(
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
						labeled, err := data.GetLabeledByLabelableAndLabel(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: labeled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type LabeledLoader struct {
	batchGet                    *dataloader.Loader
	batchGetByLabelableAndLabel *dataloader.Loader
}

func (r *LabeledLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LabeledLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByLabelableAndLabel.ClearAll()
}

func (r *LabeledLoader) Get(
	ctx context.Context,
	id int32,
) (*data.Labeled, error) {
	key := strconv.Itoa(int(id))
	labeledData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	labeled, ok := labeledData.(*data.Labeled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(labeled.LabelableID.String, labeled.LabelID.String)
	r.batchGetByLabelableAndLabel.Prime(ctx, compositeKey, labeled)

	return labeled, nil
}

func (r *LabeledLoader) GetByLabelableAndLabel(
	ctx context.Context,
	labelableID,
	userID string,
) (*data.Labeled, error) {
	compositeKey := newCompositeKey(labelableID, userID)
	labeledData, err := r.batchGetByLabelableAndLabel.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	labeled, ok := labeledData.(*data.Labeled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(labeled.ID.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), labeled)

	return labeled, nil
}
