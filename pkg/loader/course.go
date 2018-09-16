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

func NewCourseLoader() *CourseLoader {
	return &CourseLoader{
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
						course, err := data.GetCourse(db, key.String())
						results[i] = &dataloader.Result{Data: course, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByName: createLoader(
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
						course, err := data.GetCourseByName(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: course, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByNumber: createLoader(
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
						course, err := data.GetCourseByNumber(db, ks[0], int32(number))
						results[i] = &dataloader.Result{Data: course, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByStudyAndName: createLoader(
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
						course, err := data.GetCourseByStudyAndName(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: course, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type CourseLoader struct {
	batchGet               *dataloader.Loader
	batchGetByName         *dataloader.Loader
	batchGetByNumber       *dataloader.Loader
	batchGetByStudyAndName *dataloader.Loader
}

func (r *CourseLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *CourseLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByName.ClearAll()
	r.batchGetByNumber.ClearAll()
	r.batchGetByStudyAndName.ClearAll()
}

func (r *CourseLoader) Get(
	ctx context.Context,
	id string,
) (*data.Course, error) {
	courseData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	course, ok := courseData.(*data.Course)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return course, nil
}

func (r *CourseLoader) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*data.Course, error) {
	compositeKey := newCompositeKey(studyID, name)
	courseData, err := r.batchGetByName.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	course, ok := courseData.(*data.Course)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(course.ID.String), course)

	return course, nil
}

func (r *CourseLoader) GetByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (*data.Course, error) {
	compositeKey := newCompositeKey(studyID, fmt.Sprintf("%d", number))
	courseData, err := r.batchGetByNumber.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	course, ok := courseData.(*data.Course)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(course.ID.String), course)

	return course, nil
}

func (r *CourseLoader) GetByStudyAndName(
	ctx context.Context,
	study,
	name string,
) (*data.Course, error) {
	compositeKey := newCompositeKey(study, name)
	courseData, err := r.batchGetByStudyAndName.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	course, ok := courseData.(*data.Course)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(course.ID.String), course)

	return course, nil
}

func (r *CourseLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Course, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	courseData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	courses := make([]*data.Course, len(courseData))
	for i, d := range courseData {
		var ok bool
		courses[i], ok = d.(*data.Course)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return courses, nil
}
