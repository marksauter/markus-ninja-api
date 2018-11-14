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

func NewUserLoader() *UserLoader {
	return &UserLoader{
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
						user, err := data.ExistsUser(db, key.String())
						results[i] = &dataloader.Result{Data: user, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchExistsByLogin: createLoader(
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
						user, err := data.ExistsUserByLogin(db, key.String())
						results[i] = &dataloader.Result{Data: user, Error: err}
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
						user, err := data.GetUser(db, key.String())
						results[i] = &dataloader.Result{Data: user, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByLogin: createLoader(
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
						user, err := data.GetUserByLogin(db, key.String())
						results[i] = &dataloader.Result{Data: user, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type UserLoader struct {
	batchExists        *dataloader.Loader
	batchExistsByLogin *dataloader.Loader
	batchGet           *dataloader.Loader
	batchGetByLogin    *dataloader.Loader
}

func (r *UserLoader) Clear(id string) {
	ctx := context.Background()
	r.batchExists.Clear(ctx, dataloader.StringKey(id))
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserLoader) ClearAll() {
	r.batchExists.ClearAll()
	r.batchExistsByLogin.ClearAll()
	r.batchGet.ClearAll()
	r.batchGetByLogin.ClearAll()
}

func (r *UserLoader) Exists(
	ctx context.Context,
	id string,
) (bool, error) {
	userData, err := r.batchExists.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return false, err
	}
	exists, ok := userData.(bool)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}

	return exists, nil
}

func (r *UserLoader) ExistsByLogin(
	ctx context.Context,
	login string,
) (bool, error) {
	userData, err := r.batchExistsByLogin.Load(ctx, dataloader.StringKey(login))()
	if err != nil {
		return false, err
	}
	exists, ok := userData.(bool)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return false, err
	}

	return exists, nil
}

func (r *UserLoader) Get(
	ctx context.Context,
	id string,
) (*data.User, error) {
	userData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, ok := userData.(*data.User)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGetByLogin.Prime(ctx, dataloader.StringKey(user.Login.String), user)

	return user, nil
}

func (r *UserLoader) GetByLogin(
	ctx context.Context,
	id string,
) (*data.User, error) {
	userData, err := r.batchGetByLogin.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	user, ok := userData.(*data.User)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(user.ID.String), user)

	return user, nil
}

func (r *UserLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.User, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	userData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	users := make([]*data.User, len(userData))
	for i, d := range userData {
		var ok bool
		users[i], ok = d.(*data.User)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return users, nil
}
