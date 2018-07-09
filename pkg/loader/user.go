package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewUserLoader() *UserLoader {
	return &UserLoader{
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
	batchGet        *dataloader.Loader
	batchGetByLogin *dataloader.Loader
}

func (r *UserLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserLoader) Get(
	ctx context.Context,
	id string,
) (*data.User, error) {
	userData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	user, ok := userData.(*data.User)
	if !ok {
		return nil, fmt.Errorf("wrong type")
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
		return nil, err
	}
	user, ok := userData.(*data.User)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(user.Id.String), user)

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
		return nil, errs
	}
	users := make([]*data.User, len(userData))
	for i, d := range userData {
		var ok bool
		users[i], ok = d.(*data.User)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return users, nil
}
