package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewEVTLoader(
	svc *data.EVTService,
) *EVTLoader {
	return &EVTLoader{
		svc: svc,
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
						emailVerificationToken, err := svc.GetByPK(ks[0], ks[1], ks[2])
						results[i] = &dataloader.Result{Data: emailVerificationToken, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type EVTLoader struct {
	svc *data.EVTService

	batchGet *dataloader.Loader
}

func (r *EVTLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *EVTLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *EVTLoader) Get(
	emailId,
	userId,
	token string,
) (*data.EVT, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(emailId, userId, token)
	emailVerificationTokenData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	emailVerificationToken, ok := emailVerificationTokenData.(*data.EVT)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return emailVerificationToken, nil
}
