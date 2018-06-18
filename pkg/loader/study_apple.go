package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewStudyAppleLoader(svc *data.StudyAppleService) *StudyAppleLoader {
	return &StudyAppleLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetStudyAppleBy2Fn(svc.Get)),
	}
}

type StudyAppleLoader struct {
	svc *data.StudyAppleService

	batchGet *dataloader.Loader
}

func (r *StudyAppleLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *StudyAppleLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *StudyAppleLoader) Get(studyId, userId string) (*data.StudyApple, error) {
	ctx := context.Background()
	compositeKey := newCompositeKey(studyId, userId)
	studyAppleData, err := r.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	studyApple, ok := studyAppleData.(*data.StudyApple)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return studyApple, nil
}

func (r *StudyAppleLoader) GetMany(ids *[]string) ([]*data.StudyApple, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	studyAppleData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	studyApples := make([]*data.StudyApple, len(studyAppleData))
	for i, d := range studyAppleData {
		var ok bool
		studyApples[i], ok = d.(*data.StudyApple)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return studyApples, nil
}

func newBatchGetStudyAppleBy2Fn(
	getter func(string, string) (*data.StudyApple, error),
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
				studyApple, err := getter(ks[0], ks[1])
				results[i] = &dataloader.Result{Data: studyApple, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
