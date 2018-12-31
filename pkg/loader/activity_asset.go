package loader

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewActivityAssetLoader() *ActivityAssetLoader {
	return &ActivityAssetLoader{
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
						activityAsset, err := data.GetActivityAsset(db, key.String())
						results[i] = &dataloader.Result{Data: activityAsset, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByActivityAndNumber: createLoader(
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
						number, err := strconv.ParseInt(ks[1], 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: errors.New("failed to parse activity asset number")}
							return
						}
						activityAsset, err := data.GetActivityAssetByActivityAndNumber(
							db,
							ks[0],
							int32(number),
						)
						results[i] = &dataloader.Result{Data: activityAsset, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type ActivityAssetLoader struct {
	batchGet                    *dataloader.Loader
	batchGetByActivityAndNumber *dataloader.Loader
}

func (r *ActivityAssetLoader) Clear(assetID string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(assetID))
}

func (r *ActivityAssetLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByActivityAndNumber.ClearAll()
}

func (r *ActivityAssetLoader) Get(
	ctx context.Context,
	assetID string,
) (*data.ActivityAsset, error) {
	activityAssetData, err := r.batchGet.Load(ctx, dataloader.StringKey(assetID))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activityAsset, ok := activityAssetData.(*data.ActivityAsset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return activityAsset, nil
}

func (r *ActivityAssetLoader) GetByActivityAndNumber(
	ctx context.Context,
	activityID string,
	number int32,
) (*data.ActivityAsset, error) {
	compositeKey := newCompositeKey(activityID, fmt.Sprintf("%d", number))
	activityAssetData, err := r.batchGetByActivityAndNumber.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activityAsset, ok := activityAssetData.(*data.ActivityAsset)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(activityAsset.AssetID.String), activityAsset)

	return activityAsset, nil
}
