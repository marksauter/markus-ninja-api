package connector

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/model"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

func NewUserConnector(svcs *service.Services) *UserConnector {
	return &UserConnector{
		svc: svcs.User,
		batchGet: createLoader(func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
			var (
				n       = len(keys)
				results = make([]*dataloader.Result, n)
				wg      sync.WaitGroup
			)

			wg.Add(n)

			for i, key := range keys {
				go func(i int, key dataloader.Key) {
					defer wg.Done()
					user, err := svcs.User.Get(key.String())
					results[i] = &dataloader.Result{Data: user, Error: err}
				}(i, key)
			}

			wg.Wait()

			return results
		}),
	}
}

type UserConnector struct {
	svc *service.UserService

	batchGet *dataloader.Loader
}

func (r *UserConnector) Get(id string) (*model.User, error) {
	ctx := context.Background()
	data, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	user, ok := data.(*model.User)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return user, nil
}

func (r *UserConnector) GetMany(ids *[]string) ([]*model.User, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	data, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	users := make([]*model.User, len(data))
	for i, d := range data {
		var ok bool
		users[i], ok = d.(*model.User)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return users, nil
}

func (r *UserConnector) GetByLogin(login string) (*model.User, error) {
	return r.svc.GetByLogin(login)
}

func (r *UserConnector) VerifyCredentials(userCredentials *model.UserCredentials) (*model.User, error) {
	return r.svc.VerifyCredentials(userCredentials)
}

func (r *UserConnector) Create(input *service.CreateUserInput) (*model.User, error) {
	return r.svc.Create(input)
}
