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

func NewEmailLoader() *EmailLoader {
	return &EmailLoader{
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
						email, err := data.GetEmail(db, key.String())
						results[i] = &dataloader.Result{Data: email, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByValue: createLoader(
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
						email, err := data.GetEmailByValue(db, key.String())
						results[i] = &dataloader.Result{Data: email, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type EmailLoader struct {
	batchGet        *dataloader.Loader
	batchGetByValue *dataloader.Loader
}

func (r *EmailLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *EmailLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByValue.ClearAll()
}

func (r *EmailLoader) Get(
	ctx context.Context,
	id string,
) (*data.Email, error) {
	emailData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	email, ok := emailData.(*data.Email)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGetByValue.Prime(ctx, dataloader.StringKey(email.Value.String), email)

	return email, nil
}

func (r *EmailLoader) GetByValue(
	ctx context.Context,
	value string,
) (*data.Email, error) {
	emailData, err := r.batchGetByValue.Load(ctx, dataloader.StringKey(value))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	email, ok := emailData.(*data.Email)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(email.ID.String), email)

	return email, nil
}
