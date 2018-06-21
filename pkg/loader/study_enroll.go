package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewStudyEnrollLoader(svc *data.StudyEnrollService) *StudyEnrollLoader {
	return &StudyEnrollLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetStudyEnrollBy2Fn(svc.Get)),
	}
}

type StudyEnrollLoader struct {
	svc *data.StudyEnrollService

	batchGet *dataloader.Loader
}

func (r *StudyEnrollLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *StudyEnrollLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *StudyEnrollLoader) Get(studyId, userId string) (*data.StudyEnroll, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(studyId, userId)
	studyEnrollData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	studyEnroll, ok := studyEnrollData.(*data.StudyEnroll)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return studyEnroll, nil
}

func (r *StudyEnrollLoader) GetMany(ids *[]string) ([]*data.StudyEnroll, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	studyEnrollData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	studyEnrolls := make([]*data.StudyEnroll, len(studyEnrollData))
	for i, d := range studyEnrollData {
		var ok bool
		studyEnrolls[i], ok = d.(*data.StudyEnroll)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return studyEnrolls, nil
}

func newBatchGetStudyEnrollBy2Fn(
	getter func(string, string) (*data.StudyEnroll, error),
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
				studyEnroll, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: studyEnroll, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
