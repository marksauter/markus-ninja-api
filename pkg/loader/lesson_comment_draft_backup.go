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

func NewLessonCommentDraftBackupLoader() *LessonCommentDraftBackupLoader {
	return &LessonCommentDraftBackupLoader{
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
						lessonComment, err := data.GetLessonCommentDraftBackup(db, ks[0], int32(id))
						results[i] = &dataloader.Result{Data: lessonComment, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type LessonCommentDraftBackupLoader struct {
	batchGet *dataloader.Loader
}

func (r *LessonCommentDraftBackupLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *LessonCommentDraftBackupLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *LessonCommentDraftBackupLoader) Get(
	ctx context.Context,
	lessonCommentID string,
	id int32,
) (*data.LessonCommentDraftBackup, error) {
	compositeKey := newCompositeKey(lessonCommentID, fmt.Sprintf("%d", id))
	lessonCommentData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	lessonComment, ok := lessonCommentData.(*data.LessonCommentDraftBackup)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return lessonComment, nil
}

func (r *LessonCommentDraftBackupLoader) GetMany(
	ctx context.Context,
	lessonCommentID string,
	ids *[]int32,
) ([]*data.LessonCommentDraftBackup, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, id := range *ids {
		compositeKey := newCompositeKey(lessonCommentID, fmt.Sprintf("%d", id))
		keys[i] = compositeKey
	}
	lessonCommentData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	lessonComments := make([]*data.LessonCommentDraftBackup, len(lessonCommentData))
	for i, d := range lessonCommentData {
		var ok bool
		lessonComments[i], ok = d.(*data.LessonCommentDraftBackup)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return lessonComments, nil
}
