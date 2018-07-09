package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewLessonLoader() *LessonLoader {
	return &LessonLoader{
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
						lesson, err := data.GetLesson(db, key.String())
						results[i] = &dataloader.Result{Data: lesson, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type LessonLoader struct {
	batchGet *dataloader.Loader
}

func (r *LessonLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LessonLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *LessonLoader) Get(
	ctx context.Context,
	id string,
) (*data.Lesson, error) {
	lessonData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	lesson, ok := lessonData.(*data.Lesson)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return lesson, nil
}

func (r *LessonLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Lesson, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	lessonData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	lessons := make([]*data.Lesson, len(lessonData))
	for i, d := range lessonData {
		var ok bool
		lessons[i], ok = d.(*data.Lesson)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return lessons, nil
}
