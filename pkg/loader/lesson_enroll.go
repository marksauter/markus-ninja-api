package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewLessonEnrollLoader(svc *data.LessonEnrollService) *LessonEnrollLoader {
	return &LessonEnrollLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetLessonEnrollBy2Fn(svc.Get)),
	}
}

type LessonEnrollLoader struct {
	svc *data.LessonEnrollService

	batchGet *dataloader.Loader
}

func (r *LessonEnrollLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LessonEnrollLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *LessonEnrollLoader) Get(enrollableId, userId string) (*data.LessonEnroll, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(enrollableId, userId)
	lessonEnrollData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	lessonEnroll, ok := lessonEnrollData.(*data.LessonEnroll)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return lessonEnroll, nil
}

func newBatchGetLessonEnrollBy2Fn(
	getter func(string, string) (*data.LessonEnroll, error),
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
				lessonEnroll, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: lessonEnroll, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
