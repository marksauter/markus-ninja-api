package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewLessonLoader(svc *data.LessonService) *LessonLoader {
	return &LessonLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetLessonFn(svc.GetByPK)),
	}
}

type LessonLoader struct {
	svc *data.LessonService

	batchGet        *dataloader.Loader
	batchGetByLogin *dataloader.Loader
}

func (r *LessonLoader) Get(id string) (*data.LessonModel, error) {
	ctx := context.Background()
	lessonData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	lesson, ok := lessonData.(*data.LessonModel)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return lesson, nil
}

func (r *LessonLoader) GetMany(ids *[]string) ([]*data.LessonModel, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	lessonData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	lessons := make([]*data.LessonModel, len(lessonData))
	for i, d := range lessonData {
		var ok bool
		lessons[i], ok = d.(*data.LessonModel)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return lessons, nil
}

func newBatchGetLessonFn(
	getter func(string) (*data.LessonModel, error),
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
				lesson, err := getter(key.String())
				results[i] = &dataloader.Result{Data: lesson, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
