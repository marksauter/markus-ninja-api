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
		svc:                        svc,
		batchGet:                   createLoader(newBatchGetUserAssetBy1Fn(svc.Get)),
		batchGetByName:             createLoader(newBatchGetUserAssetBy3Fn(svc.GetByName)),
		batchGetByUserStudyAndName: createLoader(newBatchGetUserAssetBy3Fn(svc.GetByUserStudyAndName)),
	}
}

type UserAssetLoader struct {
	svc *data.UserAssetService

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

	compositeKey := newCompositeKey(
		userAsset.UserId.String,
		userAsset.StudyId.String,
		userAsset.Name.String,
	)
	r.batchGetByName.Prime(ctx, compositeKey, userAsset)

	return userAsset, nil
}

func (r *UserAssetLoader) GetByName(userId, studyId, name string) (*data.UserAsset, error) {
	ctx := context.Background()
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

func (r *UserAssetLoader) GetByUserStudyAndName(userLogin, studyName, name string) (*data.UserAsset, error) {
	ctx := context.Background()
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
	compositeKey = newCompositeKey(
		userAsset.UserId.String,
		userAsset.StudyId.String,
		userAsset.Name.String,
	)
	r.batchGetByName.Prime(ctx, compositeKey, userAsset)

	return userAsset, nil
}

func newBatchGetUserAssetBy1Fn(
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

func newBatchGetUserAssetBy2Fn(
	getter func(string, string) (*data.UserAsset, error),
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
				userAsset, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: userAsset, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}

func newBatchGetUserAssetBy3Fn(
	getter func(string, string, string) (*data.UserAsset, error),
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
				userAsset, err := getter(ks[0], ks[1], ks[2])
				results[i] = &dataloader.Result{Data: userAsset, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
