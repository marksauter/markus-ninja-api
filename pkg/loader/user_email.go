package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewUserEmailLoader(svc *data.UserEmailService) *UserEmailLoader {
	return &UserEmailLoader{
		svc:             svc,
		batchGet:        createLoader(newBatchGetUserEmailFn(svc.GetByPK)),
		batchGetByEmail: createLoader(newBatchGetUserEmailFn(svc.GetByEmail)),
	}
}

type UserEmailLoader struct {
	svc *data.UserEmailService

	batchGet        *dataloader.Loader
	batchGetByEmail *dataloader.Loader
}

func (r *UserEmailLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserEmailLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserEmailLoader) Get(emailId string) (*data.UserEmail, error) {
	ctx := context.Background()
	userEmailData, err := r.batchGet.Load(ctx, dataloader.StringKey(emailId))()
	if err != nil {
		return nil, err
	}
	userEmail, ok := userEmailData.(*data.UserEmail)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGetByEmail.Prime(ctx, dataloader.StringKey(userEmail.EmailValue.String), userEmail)

	return userEmail, nil
}

func (r *UserEmailLoader) GetByEmail(email string) (*data.UserEmail, error) {
	ctx := context.Background()
	userEmailData, err := r.batchGetByEmail.Load(ctx, dataloader.StringKey(email))()
	if err != nil {
		return nil, err
	}
	userEmail, ok := userEmailData.(*data.UserEmail)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(userEmail.EmailId.String), userEmail)

	return userEmail, nil
}

func newBatchGetUserEmailFn(
	getter func(string) (*data.UserEmail, error),
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
				userEmail, err := getter(key.String())
				results[i] = &dataloader.Result{Data: userEmail, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
