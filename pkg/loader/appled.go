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

func NewAppledLoader() *AppledLoader {
	return &AppledLoader{
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
						appled, err := data.GetAppled(db, int32(id))
						results[i] = &dataloader.Result{Data: appled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByAppleableAndUser: createLoader(
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
						appled, err := data.GetAppledByAppleableAndUser(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: appled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type AppledLoader struct {
	batchGet                   *dataloader.Loader
	batchGetByAppleableAndUser *dataloader.Loader
}

func (r *AppledLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *AppledLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByAppleableAndUser.ClearAll()
}

func (r *AppledLoader) Get(ctx context.Context, id int32) (*data.Appled, error) {
	key := strconv.Itoa(int(id))
	appledData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	appled, ok := appledData.(*data.Appled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(appled.AppleableID.String, appled.UserID.String)
	r.batchGetByAppleableAndUser.Prime(ctx, compositeKey, appled)

	return appled, nil
}

func (r *AppledLoader) GetByAppleableAndUser(
	ctx context.Context,
	appleableID,
	userID string,
) (*data.Appled, error) {
	compositeKey := newCompositeKey(appleableID, userID)
	appledData, err := r.batchGetByAppleableAndUser.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	appled, ok := appledData.(*data.Appled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(appled.ID.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), appled)

	return appled, nil
}
