package loader

import (
	"context"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewAssetLoader() *AssetLoader {
	return &AssetLoader{
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
						id, err := strconv.ParseInt(key.String(), 10, 64)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						asset, err := data.GetAsset(db, id)
						results[i] = &dataloader.Result{Data: asset, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByKey: createLoader(
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
						asset, err := data.GetAssetByKey(db, key.String())
						results[i] = &dataloader.Result{Data: asset, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type AssetLoader struct {
	batchGet      *dataloader.Loader
	batchGetByKey *dataloader.Loader
}

func (r *AssetLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *AssetLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByKey.ClearAll()
}

func (r *AssetLoader) Get(
	ctx context.Context,
	id int64,
) (*data.Asset, error) {
	key := strconv.Itoa(int(id))
	assetData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	asset, ok := assetData.(*data.Asset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return asset, nil
}

func (r *AssetLoader) GetByKey(
	ctx context.Context,
	key string,
) (*data.Asset, error) {
	assetData, err := r.batchGetByKey.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	asset, ok := assetData.(*data.Asset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return asset, nil
}

func (r *AssetLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Asset, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	assetData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	assets := make([]*data.Asset, len(assetData))
	for i, d := range assetData {
		var ok bool
		assets[i], ok = d.(*data.Asset)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return assets, nil
}
