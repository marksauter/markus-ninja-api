package loader

import (
	"context"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewCommentLoader() *CommentLoader {
	return &CommentLoader{
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
						comment, err := data.GetComment(db, key.String())
						results[i] = &dataloader.Result{Data: comment, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type CommentLoader struct {
	batchGet *dataloader.Loader
}

func (r *CommentLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *CommentLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *CommentLoader) Get(
	ctx context.Context,
	id string,
) (*data.Comment, error) {
	commentData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	comment, ok := commentData.(*data.Comment)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return comment, nil
}

func (r *CommentLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Comment, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	commentData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	comments := make([]*data.Comment, len(commentData))
	for i, d := range commentData {
		var ok bool
		comments[i], ok = d.(*data.Comment)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return comments, nil
}
