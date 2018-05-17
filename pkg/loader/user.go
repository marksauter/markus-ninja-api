package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myerr"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

func NewUserLoader(svc *data.UserService) *UserLoader {
	return &UserLoader{
		svc:             svc,
		batchGet:        createLoader(newBatchGetUserFn(svc.GetById)),
		batchGetByLogin: createLoader(newBatchGetUserFn(svc.GetByLogin)),
	}
}

type UserLoader struct {
	svc *data.UserService

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

func (r *UserLoader) Get(id string) (*data.User, error) {
	ctx := context.Background()
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

func (r *UserLoader) GetMany(ids *[]string) ([]*data.User, []error) {
	ctx := context.Background()
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

func (r *UserLoader) GetByLogin(login string) (*data.User, error) {
	ctx := context.Background()
	userData, err := r.batchGetByLogin.Load(ctx, dataloader.StringKey(login))()
	if err != nil {
		return nil, err
	}
	user, ok := userData.(*data.User)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}
	id, ok := user.Id.Get().(oid.OID)
	if !ok {
		return nil, myerr.UnexpectedError{"user missing `id` field"}
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(id.String), user)

	return user, nil
}

func newBatchGetUserFn(
	getter func(string) (*data.User, error),
) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		var (
			n       = len(keys)
			results = make([]*dataloader.Result, n)
			wg      sync.WaitGroup
		)

		wg.Add(n)

		for i, key := range keys {
			go func(i int, key dataloader.Key) {
				defer wg.Done()
				user, err := getter(key.String())
				results[i] = &dataloader.Result{Data: user, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
