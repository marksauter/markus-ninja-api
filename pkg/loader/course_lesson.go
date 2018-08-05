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
)

func NewCourseLessonLoader() *CourseLessonLoader {
	return &CourseLessonLoader{
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
						courseLesson, err := data.GetCourseLesson(db, key.String())
						results[i] = &dataloader.Result{Data: courseLesson, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByCourseAndNumber: createLoader(
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
							results[i] = &dataloader.Result{Error: errors.New("failed to parse course lesson number")}
							return
						}
						courseLesson, err := data.GetCourseLessonByCourseAndNumber(
							db,
							ks[0],
							int32(number),
						)
						results[i] = &dataloader.Result{Data: courseLesson, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type CourseLessonLoader struct {
	batchGet                  *dataloader.Loader
	batchGetByCourseAndNumber *dataloader.Loader
}

func (r *CourseLessonLoader) Clear(lessonId string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(lessonId))
}

func (r *CourseLessonLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByCourseAndNumber.ClearAll()
}

func (r *CourseLessonLoader) Get(
	ctx context.Context,
	lessonId string,
) (*data.CourseLesson, error) {
	courseLessonData, err := r.batchGet.Load(ctx, dataloader.StringKey(lessonId))()
	if err != nil {
		return nil, err
	}
	courseLesson, ok := courseLessonData.(*data.CourseLesson)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return courseLesson, nil
}

func (r *CourseLessonLoader) GetByCourseAndNumber(
	ctx context.Context,
	courseId string,
	number int32,
) (*data.CourseLesson, error) {
	compositeKey := newCompositeKey(courseId, fmt.Sprintf("%d", number))
	courseLessonData, err := r.batchGetByCourseAndNumber.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	courseLesson, ok := courseLessonData.(*data.CourseLesson)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(courseLesson.LessonId.String), courseLesson)

	return courseLesson, nil
}
