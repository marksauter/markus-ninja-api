package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewUserAssetCommentLoader() *UserAssetCommentLoader {
	return &UserAssetCommentLoader{
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
						userAssetComment, err := data.GetUserAssetComment(db, key.String())
						results[i] = &dataloader.Result{Data: userAssetComment, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type UserAssetCommentLoader struct {
	batchGet *dataloader.Loader
}

func (r *UserAssetCommentLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *UserAssetCommentLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *UserAssetCommentLoader) Get(
	ctx context.Context,
	id string,
) (*data.UserAssetComment, error) {
	userAssetCommentData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	userAssetComment, ok := userAssetCommentData.(*data.UserAssetComment)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return userAssetComment, nil
}

func (r *UserAssetCommentLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.UserAssetComment, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	userAssetCommentData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	userAssetComments := make([]*data.UserAssetComment, len(userAssetCommentData))
	for i, d := range userAssetCommentData {
		var ok bool
		userAssetComments[i], ok = d.(*data.UserAssetComment)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return userAssetComments, nil
}
