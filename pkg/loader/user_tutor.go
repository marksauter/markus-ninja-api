package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewUserTutorLoader(svc *data.UserTutorService) *UserTutorLoader {
	return &UserTutorLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetUserTutorBy2Fn(svc.Get)),
	}
}

type UserTutorLoader struct {
	svc *data.UserTutorService

	batchGet *dataloader.Loader
}

func (r *UserTutorLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserTutorLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserTutorLoader) Get(tutorId, pupilId string) (*data.UserTutor, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(tutorId, pupilId)
	userTutorData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userTutor, ok := userTutorData.(*data.UserTutor)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return userTutor, nil
}

func newBatchGetUserTutorBy2Fn(
	getter func(string, string) (*data.UserTutor, error),
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
				userTutor, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: userTutor, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
