package svccxn

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func NewUserConnection(svc *service.UserService) *UserConnection {
	return &UserConnection{
		svc:             svc,
		batchGet:        createLoader(newBatchGetFn(svc.Get)),
		batchGetByLogin: createLoader(newBatchGetFn(svc.GetByLogin)),
	}
}

type UserConnection struct {
	svc *service.UserService

	batchGet        *dataloader.Loader
	batchGetByLogin *dataloader.Loader
}

func (r *UserConnection) Create(
	input *service.CreateUserInput,
) (*service.UserModel, error) {
	return r.svc.Create(input)
}

func (r *UserConnection) Get(id string) (*service.UserModel, error) {
	ctx := context.Background()
	data, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	user, ok := data.(*service.UserModel)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGetByLogin.Prime(ctx, dataloader.StringKey(user.Login), user)

	return user, nil
}

func (r *UserConnection) GetMany(ids *[]string) ([]*service.UserModel, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	data, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	users := make([]*service.UserModel, len(data))
	for i, d := range data {
		var ok bool
		users[i], ok = d.(*service.UserModel)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return users, nil
}

func (r *UserConnection) GetByLogin(login string) (*service.UserModel, error) {
	ctx := context.Background()
	data, err := r.batchGetByLogin.Load(ctx, dataloader.StringKey(login))()
	if err != nil {
		return nil, err
	}
	user, ok := data.(*service.UserModel)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(user.Id), user)

	return user, nil
}

func (r *UserConnection) VerifyCredentials(
	input *service.VerifyCredentialsInput,
) (*service.UserModel, error) {
	return r.svc.VerifyCredentials(input)
}

func newBatchGetFn(
	getter func(string) (*service.UserModel, error),
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
