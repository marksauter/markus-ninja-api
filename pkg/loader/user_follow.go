package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewUserFollowLoader(svc *data.UserFollowService) *UserFollowLoader {
	return &UserFollowLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetUserFollowBy2Fn(svc.Get)),
	}
}

type UserFollowLoader struct {
	svc *data.UserFollowService

	batchGet *dataloader.Loader
}

func (r *UserFollowLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserFollowLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserFollowLoader) Get(leaderId, followerId string) (*data.UserFollow, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(leaderId, followerId)
	userFollowData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userFollow, ok := userFollowData.(*data.UserFollow)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return userFollow, nil
}

func newBatchGetUserFollowBy2Fn(
	getter func(string, string) (*data.UserFollow, error),
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
				userFollow, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: userFollow, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
