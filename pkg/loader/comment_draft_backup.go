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

func NewCommentDraftBackupLoader() *CommentDraftBackupLoader {
	return &CommentDraftBackupLoader{
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
						comment, err := data.GetCommentDraftBackup(db, ks[0], int32(id))
						results[i] = &dataloader.Result{Data: comment, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type CommentDraftBackupLoader struct {
	batchGet *dataloader.Loader
}

func (r *CommentDraftBackupLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *CommentDraftBackupLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *CommentDraftBackupLoader) Get(
	ctx context.Context,
	commentID string,
	id int32,
) (*data.CommentDraftBackup, error) {
	compositeKey := newCompositeKey(commentID, fmt.Sprintf("%d", id))
	commentData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	comment, ok := commentData.(*data.CommentDraftBackup)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return comment, nil
}

func (r *CommentDraftBackupLoader) GetMany(
	ctx context.Context,
	commentID string,
	ids *[]int32,
) ([]*data.CommentDraftBackup, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, id := range *ids {
		compositeKey := newCompositeKey(commentID, fmt.Sprintf("%d", id))
		keys[i] = compositeKey
	}
	commentData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	comments := make([]*data.CommentDraftBackup, len(commentData))
	for i, d := range commentData {
		var ok bool
		comments[i], ok = d.(*data.CommentDraftBackup)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return comments, nil
}
