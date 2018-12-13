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

func NewLessonDraftBackupLoader() *LessonDraftBackupLoader {
	return &LessonDraftBackupLoader{
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
						ks := splitCompositeKey(key)
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						id, err := strconv.ParseInt(ks[1], 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						lesson, err := data.GetLessonDraftBackup(db, ks[0], int32(id))
						results[i] = &dataloader.Result{Data: lesson, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type LessonDraftBackupLoader struct {
	batchGet *dataloader.Loader
}

func (r *LessonDraftBackupLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LessonDraftBackupLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *LessonDraftBackupLoader) Get(
	ctx context.Context,
	lessonID string,
	id int32,
) (*data.LessonDraftBackup, error) {
	compositeKey := newCompositeKey(lessonID, fmt.Sprintf("%d", id))
	lessonData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lesson, ok := lessonData.(*data.LessonDraftBackup)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return lesson, nil
}

func (r *LessonDraftBackupLoader) GetMany(
	ctx context.Context,
	lessonID string,
	ids *[]int32,
) ([]*data.LessonDraftBackup, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, id := range *ids {
		compositeKey := newCompositeKey(lessonID, fmt.Sprintf("%d", id))
		keys[i] = compositeKey
	}
	lessonData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	lessons := make([]*data.LessonDraftBackup, len(lessonData))
	for i, d := range lessonData {
		var ok bool
		lessons[i], ok = d.(*data.LessonDraftBackup)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return lessons, nil
}
