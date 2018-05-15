package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewStudyLoader(svc *data.StudyService) *StudyLoader {
	return &StudyLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetStudyFn(svc.GetByPK)),
	}
}

type StudyLoader struct {
	svc *data.StudyService

	batchGet        *dataloader.Loader
	batchGetByLogin *dataloader.Loader
}

func (r *StudyLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *StudyLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *StudyLoader) Get(id string) (*data.Study, error) {
	ctx := context.Background()
	studyData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	study, ok := studyData.(*data.Study)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return study, nil
}

func (r *StudyLoader) GetMany(ids *[]string) ([]*data.Study, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	studyData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	studys := make([]*data.Study, len(studyData))
	for i, d := range studyData {
		var ok bool
		studys[i], ok = d.(*data.Study)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return studys, nil
}

func newBatchGetStudyFn(
	getter func(string) (*data.Study, error),
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
				study, err := getter(key.String())
				results[i] = &dataloader.Result{Data: study, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
