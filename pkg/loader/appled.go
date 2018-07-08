package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewAppledLoader(db data.Queryer) *AppledLoader {
	return &AppledLoader{
		db: db,
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
						appled, err := data.GetAppled(db, int32(id))
						results[i] = &dataloader.Result{Data: appled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetForAppleable: createLoader(
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
						appled, err := data.GetAppledForAppleable(db, ks[0], ks[1])
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
	db                   data.Queryer
	batchGet             *dataloader.Loader
	batchGetForAppleable *dataloader.Loader
}

func (r *AppledLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *AppledLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetForAppleable.ClearAll()
}

func (r *AppledLoader) Get(id int32) (*data.Appled, error) {
	ctx := context.Background()
	key := strconv.Itoa(int(id))
	appledData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	appled, ok := appledData.(*data.Appled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(appled.AppleableId.String, appled.UserId.String)
	r.batchGetForAppleable.Prime(ctx, compositeKey, appled)

	return appled, nil
}

func (r *AppledLoader) GetForAppleable(appleableId, userId string) (*data.Appled, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(appleableId, userId)
	appledData, err := r.batchGetForAppleable.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	appled, ok := appledData.(*data.Appled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(appled.Id.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), appled)

	return appled, nil
}
