package loader

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewActivityLoader() *ActivityLoader {
	return &ActivityLoader{
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
						activity, err := data.GetActivity(db, key.String())
						results[i] = &dataloader.Result{Data: activity, Error: err}
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
						activity, err := data.GetActivityByName(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: activity, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByNumber: createLoader(
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
						number, err := strconv.ParseInt(ks[1], 10, 32)
						if err != nil {
							results[i] = &dataloader.Result{Error: errors.New("failed to parse activity lesson number")}
							return
						}
						activity, err := data.GetActivityByNumber(db, ks[0], int32(number))
						results[i] = &dataloader.Result{Data: activity, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
		batchGetByStudyAndName: createLoader(
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
						activity, err := data.GetActivityByStudyAndName(db, ks[0], ks[1])
						results[i] = &dataloader.Result{Data: activity, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type ActivityLoader struct {
	batchGet               *dataloader.Loader
	batchGetByName         *dataloader.Loader
	batchGetByNumber       *dataloader.Loader
	batchGetByStudyAndName *dataloader.Loader
}

func (r *ActivityLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *ActivityLoader) ClearAll() {
	r.batchGet.ClearAll()
	r.batchGetByName.ClearAll()
	r.batchGetByNumber.ClearAll()
	r.batchGetByStudyAndName.ClearAll()
}

func (r *ActivityLoader) Get(
	ctx context.Context,
	id string,
) (*data.Activity, error) {
	activityData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, ok := activityData.(*data.Activity)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return activity, nil
}

func (r *ActivityLoader) GetByName(
	ctx context.Context,
	studyID,
	name string,
) (*data.Activity, error) {
	compositeKey := newCompositeKey(studyID, name)
	activityData, err := r.batchGetByName.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, ok := activityData.(*data.Activity)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(activity.ID.String), activity)

	return activity, nil
}

func (r *ActivityLoader) GetByNumber(
	ctx context.Context,
	studyID string,
	number int32,
) (*data.Activity, error) {
	compositeKey := newCompositeKey(studyID, fmt.Sprintf("%d", number))
	activityData, err := r.batchGetByNumber.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, ok := activityData.(*data.Activity)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(activity.ID.String), activity)

	return activity, nil
}

func (r *ActivityLoader) GetByStudyAndName(
	ctx context.Context,
	study,
	name string,
) (*data.Activity, error) {
	compositeKey := newCompositeKey(study, name)
	activityData, err := r.batchGetByStudyAndName.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	activity, ok := activityData.(*data.Activity)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	r.batchGet.Prime(ctx, dataloader.StringKey(activity.ID.String), activity)

	return activity, nil
}

func (r *ActivityLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Activity, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	activityData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	activities := make([]*data.Activity, len(activityData))
	for i, d := range activityData {
		var ok bool
		activities[i], ok = d.(*data.Activity)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return activities, nil
}
