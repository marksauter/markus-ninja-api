package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
)

func NewEnrolledLoader() *EnrolledLoader {
	return &EnrolledLoader{
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
						id, err := strconv.ParseInt(key.String(), 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: err}
							return
						}
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						enrolled, err := data.GetEnrolled(db, int32(id))
						results[i] = &dataloader.Result{Data: enrolled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByEnrollableAndUser: createLoader(
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
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						enrolled, err := data.GetEnrolledByEnrollableAndUser(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: enrolled, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type EnrolledLoader struct {
	batchGet                    *dataloader.Loader
	batchGetByEnrollableAndUser *dataloader.Loader
}

func (r *EnrolledLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *EnrolledLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByEnrollableAndUser.ClearAll()
}

func (r *EnrolledLoader) Get(
	ctx context.Context,
	id int32,
) (*data.Enrolled, error) {
	key := strconv.Itoa(int(id))
	enrolledData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	enrolled, ok := enrolledData.(*data.Enrolled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(enrolled.EnrollableID.String, enrolled.UserID.String)
	r.batchGetByEnrollableAndUser.Prime(ctx, compositeKey, enrolled)

	return enrolled, nil
}

func (r *EnrolledLoader) GetByEnrollableAndUser(
	ctx context.Context,
	enrollableID,
	userID string,
) (*data.Enrolled, error) {
	compositeKey := newCompositeKey(enrollableID, userID)
	enrolledData, err := r.batchGetByEnrollableAndUser.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	enrolled, ok := enrolledData.(*data.Enrolled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(enrolled.ID.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), enrolled)

	return enrolled, nil
}

func (r *EnrolledLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Enrolled, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	enrolledData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	enrolleds := make([]*data.Enrolled, len(enrolledData))
	for i, d := range enrolledData {
		var ok bool
		enrolleds[i], ok = d.(*data.Enrolled)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return enrolleds, nil
}
