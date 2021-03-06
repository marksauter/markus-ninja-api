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
						userAsset, err := data.GetUserAssetByName(db, ks[0], ks[1])
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
	r.batchGetByUserStudyAndName.ClearAll()
}

func (r *UserAssetLoader) Get(
	ctx context.Context,
	id string,
) (*data.UserAsset, error) {
	userAssetData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, ok := userAssetData.(*data.UserAsset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	compositeKey := newCompositeKey(userAsset.StudyID.String, userAsset.Name.String)
	r.batchGetByName.Prime(ctx, compositeKey, userAsset)

	return userAsset, nil
}

func (r *UserAssetLoader) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*data.UserAsset, error) {
	compositeKey := newCompositeKey(studyID, name)
	userAssetData, err := r.batchGetByName.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, ok := userAssetData.(*data.UserAsset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(userAsset.ID.String), userAsset)

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
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	userAsset, ok := userAssetData.(*data.UserAsset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(userAsset.ID.String), userAsset)
	compositeKey = newCompositeKey(userAsset.StudyID.String, userAsset.Name.String)
	r.batchGetByName.Prime(ctx, compositeKey, userAsset)

	return userAsset, nil
}

func (r *UserAssetLoader) GetMany(
	ctx context.Context,
	ids []string,
) ([]*data.UserAsset, []error) {
	keys := make(dataloader.Keys, len(ids))
	for i, k := range ids {
		keys[i] = dataloader.StringKey(k)
	}
	userAssetData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	userAssets := make([]*data.UserAsset, len(userAssetData))
	for i, d := range userAssetData {
		var ok bool
		userAssets[i], ok = d.(*data.UserAsset)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return userAssets, nil
}

func (r *UserAssetLoader) GetManyByName(
	ctx context.Context,
	studyID string,
	names []string,
) ([]*data.UserAsset, []error) {
	keys := make(dataloader.Keys, len(names))
	for i, name := range names {
		keys[i] = newCompositeKey(studyID, name)
	}
	userAssetData, errs := r.batchGetByName.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	userAssets := make([]*data.UserAsset, len(userAssetData))
	for i, d := range userAssetData {
		var ok bool
		userAssets[i], ok = d.(*data.UserAsset)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return userAssets, nil
}
