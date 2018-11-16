package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewLessonLoader() *LessonLoader {
	return &LessonLoader{
		batchExists: createLoader(
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
						exists, err := data.ExistsLesson(db, key.String())
						results[i] = &dataloader.Result{Data: exists, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchExistsByNumber: createLoader(
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
							results[i] = &dataloader.Result{Error: err}
							return
						}
						exists, err := data.ExistsLessonByNumber(db, ks[0], int32(number))
						results[i] = &dataloader.Result{Data: exists, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchExistsByOwnerStudyAndNumber: createLoader(
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
						number, err := strconv.ParseInt(ks[2], 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						exists, err := data.ExistsLessonByOwnerStudyAndNumber(db, ks[0], ks[1], int32(number))
						results[i] = &dataloader.Result{Data: exists, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
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
							results[i] = &dataloader.Result{Error: err}
							return
						}
						lesson, err := data.GetLessonByNumber(db, ks[0], int32(number))
						results[i] = &dataloader.Result{Data: lesson, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByOwnerStudyAndNumber: createLoader(
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
						number, err := strconv.ParseInt(ks[2], 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						lesson, err := data.GetLessonByOwnerStudyAndNumber(db, ks[0], ks[1], int32(number))
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
	batchExists                      *dataloader.Loader
	batchExistsByNumber              *dataloader.Loader
	batchExistsByOwnerStudyAndNumber *dataloader.Loader
	batchGet                         *dataloader.Loader
	batchGetByNumber                 *dataloader.Loader
	batchGetByOwnerStudyAndNumber    *dataloader.Loader
}

func (r *LessonLoader) Clear(id string) {
	ctx := context.Background()
	r.batchExists.Clear(ctx, dataloader.StringKey(id))
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LessonLoader) ClearAll() {
	r.batchExists.ClearAll()
	r.batchExistsByNumber.ClearAll()
	r.batchExistsByOwnerStudyAndNumber.ClearAll()
	r.batchGet.ClearAll()
	r.batchGetByNumber.ClearAll()
	r.batchGetByOwnerStudyAndNumber.ClearAll()
}

func (r *LessonLoader) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	lessonData, err := r.batchExists.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	exists, ok := lessonData.(bool)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}

	return exists, nil
}

func (r *LessonLoader) ExistsByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (bool, error) {
	compositeKey := newCompositeKey(studyID, fmt.Sprintf("%d", number))
	lessonData, err := r.batchExistsByNumber.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	exists, ok := lessonData.(bool)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}

	return exists, nil
}

func (r *LessonLoader) ExistsLessonByOwnerStudyAndNumber(
	ctx context.Context,
	ownerLogin,
	studyName string,
	number int32,
) (bool, error) {
	compositeKey := newCompositeKey(ownerLogin, studyName, fmt.Sprintf("%d", number))
	lessonData, err := r.batchExistsByOwnerStudyAndNumber.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}
	exists, ok := lessonData.(bool)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}

	return exists, nil
}

func (r *LessonLoader) Get(
	ctx context.Context,
	id string,
) (*data.Lesson, error) {
	lessonData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, ok := lessonData.(*data.Lesson)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	numberKey := fmt.Sprintf("%d", lesson.Number.Int)
	compositeKey := newCompositeKey(lesson.StudyID.String, numberKey)
	r.batchGetByNumber.Prime(ctx, compositeKey, lesson)

	return lesson, nil
}

func (r *LessonLoader) GetByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (*data.Lesson, error) {
	compositeKey := newCompositeKey(studyID, fmt.Sprintf("%d", number))
	lessonData, err := r.batchGetByNumber.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, ok := lessonData.(*data.Lesson)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(lesson.ID.String), lesson)

	return lesson, nil
}

func (r *LessonLoader) GetLessonByOwnerStudyAndNumber(
	ctx context.Context,
	ownerLogin,
	studyName string,
	number int32,
) (*data.Lesson, error) {
	compositeKey := newCompositeKey(ownerLogin, studyName, fmt.Sprintf("%d", number))
	lessonData, err := r.batchGetByOwnerStudyAndNumber.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, ok := lessonData.(*data.Lesson)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(lesson.ID.String), lesson)
	numberKey := fmt.Sprintf("%d", lesson.Number.Int)
	compositeKey = newCompositeKey(lesson.StudyID.String, numberKey)
	r.batchGetByNumber.Prime(ctx, compositeKey, lesson)

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
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	lessons := make([]*data.Lesson, len(lessonData))
	for i, d := range lessonData {
		var ok bool
		lessons[i], ok = d.(*data.Lesson)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return lessons, nil
}
