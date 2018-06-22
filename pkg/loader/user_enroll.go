package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewUserEnrollLoader(svc *data.UserEnrollService) *UserEnrollLoader {
	return &UserEnrollLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetUserEnrollBy2Fn(svc.Get)),
	}
}

type UserEnrollLoader struct {
	svc *data.UserEnrollService

	batchGet *dataloader.Loader
}

func (r *UserEnrollLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserEnrollLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserEnrollLoader) Get(tutorId, pupilId string) (*data.UserEnroll, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(tutorId, pupilId)
	userEnrollData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userEnroll, ok := userEnrollData.(*data.UserEnroll)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return userEnroll, nil
}

func newBatchGetUserEnrollBy2Fn(
	getter func(string, string) (*data.UserEnroll, error),
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
				userEnroll, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: userEnroll, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
