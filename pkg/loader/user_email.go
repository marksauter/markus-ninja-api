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
		svc:                      svc,
		batchGet:                 createLoader(newBatchGetUserEmailFn(svc.GetByPK)),
		batchGetByUserIdAndEmail: createLoader(newBatchGetUserEmailFn(svc.GetByUserIdAndEmail)),
	}
}

type UserEmailLoader struct {
	svc *data.UserEmailService

	batchGet                 *dataloader.Loader
	batchGetByUserIdAndEmail *dataloader.Loader
}

func (r *UserEmailLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserEmailLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserEmailLoader) Get(userId, emailId string) (*data.UserEmail, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(userId, emailId)
	userEmailData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userEmail, ok := userEmailData.(*data.UserEmail)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return userEmail, nil
}

func (r *UserEmailLoader) GetByUserIdAndEmail(userId, email string) (*data.UserEmail, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(userId, email)
	userEmailData, err := r.batchGetByUserIdAndEmail.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	userEmail, ok := userEmailData.(*data.UserEmail)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	r.batchGet.Prime(ctx, newCompositeKey(userId, userEmail.EmailId.String), userEmail)

	return userEmail, nil
}

func newBatchGetUserEmailFn(
	getter func(string, string) (*data.UserEmail, error),
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
				ks := splitCompositeKey(key)
				userEmail, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: userEmail, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
