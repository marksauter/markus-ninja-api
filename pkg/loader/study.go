package loader

import (
	"context"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewStudyLoader() *StudyLoader {
	return &StudyLoader{
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
						study, err := data.GetStudy(db, key.String())
						results[i] = &dataloader.Result{Data: study, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByName: createLoader(
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
						study, err := data.GetStudyByName(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: study, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByUserAndName: createLoader(
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
						study, err := data.GetStudyByUserAndName(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: study, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type StudyLoader struct {
	batchGet              *dataloader.Loader
	batchGetByName        *dataloader.Loader
	batchGetByUserAndName *dataloader.Loader
}

func (r *StudyLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *StudyLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByName.ClearAll()
	r.batchGetByUserAndName.ClearAll()
}

func (r *StudyLoader) Get(
	ctx context.Context,
	id string,
) (*data.Study, error) {
	studyData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	study, ok := studyData.(*data.Study)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return study, nil
}

func (r *StudyLoader) GetByName(
	ctx context.Context,
	userID,
	name string,
) (*data.Study, error) {
	compositeKey := newCompositeKey(userID, name)
	studyData, err := r.batchGetByName.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	study, ok := studyData.(*data.Study)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(study.ID.String), study)

	return study, nil
}

func (r *StudyLoader) GetByUserAndName(
	ctx context.Context,
	login,
	name string,
) (*data.Study, error) {
	compositeKey := newCompositeKey(login, name)
	studyData, err := r.batchGetByUserAndName.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	study, ok := studyData.(*data.Study)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(study.ID.String), study)

	return study, nil
}

func (r *StudyLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Study, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	studyData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	studys := make([]*data.Study, len(studyData))
	for i, d := range studyData {
		var ok bool
		studys[i], ok = d.(*data.Study)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return studys, nil
}
