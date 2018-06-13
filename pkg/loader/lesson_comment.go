package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewLessonCommentLoader(svc *data.LessonCommentService) *LessonCommentLoader {
	return &LessonCommentLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetLessonCommentFn(svc.Get)),
	}
}

type LessonCommentLoader struct {
	svc *data.LessonCommentService

	batchGet        *dataloader.Loader
	batchGetByLogin *dataloader.Loader
}

func (r *LessonCommentLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LessonCommentLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *LessonCommentLoader) Get(id string) (*data.LessonComment, error) {
	ctx := context.Background()
	lessonCommentData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	lessonComment, ok := lessonCommentData.(*data.LessonComment)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return lessonComment, nil
}

func (r *LessonCommentLoader) GetMany(ids *[]string) ([]*data.LessonComment, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	lessonCommentData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	lessonComments := make([]*data.LessonComment, len(lessonCommentData))
	for i, d := range lessonCommentData {
		var ok bool
		lessonComments[i], ok = d.(*data.LessonComment)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return lessonComments, nil
}

func newBatchGetLessonCommentFn(
	getter func(string) (*data.LessonComment, error),
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
				lessonComment, err := getter(key.String())
				results[i] = &dataloader.Result{Data: lessonComment, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
