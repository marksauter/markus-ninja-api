package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewUserAssetLoader(svc *data.UserAssetService) *UserAssetLoader {
	return &UserAssetLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetUserAssetFn(svc.GetByPK)),
	}
}

type UserAssetLoader struct {
	svc *data.UserAssetService

	batchGet *dataloader.Loader
}

func (r *UserAssetLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserAssetLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserAssetLoader) Get(id string) (*data.UserAsset, error) {
	ctx := context.Background()
	userAssetData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	userAsset, ok := userAssetData.(*data.UserAsset)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return userAsset, nil
}

func newBatchGetUserAssetFn(
	getter func(string) (*data.UserAsset, error),
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
				userAsset, err := getter(key.String())
				results[i] = &dataloader.Result{Data: userAsset, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
