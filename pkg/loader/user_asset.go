package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewUserAssetLoader() *UserAssetLoader {
	return &UserAssetLoader{
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
						userAsset, err := data.GetUserAsset(db, key.String())
						results[i] = &dataloader.Result{Data: userAsset, Error: err}
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
						ks := splitCompositeKey(key)
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						userAsset, err := data.GetUserAssetByName(db, ks[0], ks[1], ks[2])
						results[i] = &dataloader.Result{Data: userAsset, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByUserStudyAndName: createLoader(
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
						userAsset, err := data.GetUserAssetByUserStudyAndName(db, ks[0], ks[1], ks[2])
						results[i] = &dataloader.Result{Data: userAsset, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type UserAssetLoader struct {
	batchGet                   *dataloader.Loader
	batchGetByName             *dataloader.Loader
	batchGetByUserStudyAndName *dataloader.Loader
}

func (r *UserAssetLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserAssetLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByName.ClearAll()
}

func (r *UserAssetLoader) Get(
	ctx context.Context,
	id string,
) (*data.UserAsset, error) {
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

func (r *UserAssetLoader) GetByName(
	ctx context.Context,
	userId,
	studyId,
	name string,
) (*data.UserAsset, error) {
	compositeKey := newCompositeKey(userId, studyId, name)
	userAssetData, err := r.batchGetByName.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userAsset, ok := userAssetData.(*data.UserAsset)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(userAsset.Id.String), userAsset)

	return userAsset, nil
}

func (r *UserAssetLoader) GetByUserStudyAndName(
	ctx context.Context,
	userLogin,
	studyName,
	name string,
) (*data.UserAsset, error) {
	compositeKey := newCompositeKey(userLogin, studyName, name)
	userAssetData, err := r.batchGetByUserStudyAndName.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userAsset, ok := userAssetData.(*data.UserAsset)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(userAsset.Id.String), userAsset)

	return userAsset, nil
}

func (r *UserAssetLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.UserAsset, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	userAssetData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	userAssets := make([]*data.UserAsset, len(userAssetData))
	for i, d := range userAssetData {
		var ok bool
		userAssets[i], ok = d.(*data.UserAsset)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return userAssets, nil
}
