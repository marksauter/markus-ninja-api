package loader

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewEnrolledLoader(svc *data.EnrolledService) *EnrolledLoader {
	return &EnrolledLoader{
		svc:                   svc,
		batchGet:              createLoader(newBatchGetEnrolledBy1Fn(svc.Get)),
		batchGetForEnrollable: createLoader(newBatchGetEnrolledBy2Fn(svc.GetForEnrollable)),
	}
}

type EnrolledLoader struct {
	svc *data.EnrolledService

	batchGet              *dataloader.Loader
	batchGetForEnrollable *dataloader.Loader
}

func (r *EnrolledLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *EnrolledLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *EnrolledLoader) Get(id int32) (*data.Enrolled, error) {
	ctx := context.Background()
	key := strconv.Itoa(int(id))
	enrolledData, err := r.batchGet.Load(ctx, dataloader.StringKey(key))()
	if err != nil {
		return nil, err
	}
	enrolled, ok := enrolledData.(*data.Enrolled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	compositeKey := newCompositeKey(enrolled.EnrollableId.String, enrolled.UserId.String)
	r.batchGetForEnrollable.Prime(ctx, compositeKey, enrolled)

	return enrolled, nil
}

func (r *EnrolledLoader) GetForEnrollable(enrollableId, userId string) (*data.Enrolled, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(enrollableId, userId)
	enrolledData, err := r.batchGetForEnrollable.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	enrolled, ok := enrolledData.(*data.Enrolled)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	key := strconv.Itoa(int(enrolled.Id.Int))
	r.batchGet.Prime(ctx, dataloader.StringKey(key), enrolled)

	return enrolled, nil
}

func (r *EnrolledLoader) GetMany(ids *[]string) ([]*data.Enrolled, []error) {
	ctx := context.Background()
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

func newBatchGetEnrolledBy1Fn(
	getter func(int32) (*data.Enrolled, error),
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
				id, err := strconv.ParseInt(key.String(), 10, 32)
				if err != nil {
					results[i] = &dataloader.Result{Error: err}
					return
				}
				enrolled, err := getter(int32(id))
				results[i] = &dataloader.Result{Data: enrolled, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}

func newBatchGetEnrolledBy2Fn(
	getter func(string, string) (*data.Enrolled, error),
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
				enrolled, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: enrolled, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
