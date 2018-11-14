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

func NewNotificationLoader() *NotificationLoader {
	return &NotificationLoader{
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
						notification, err := data.GetNotification(db, key.String())
						results[i] = &dataloader.Result{Data: notification, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type NotificationLoader struct {
	batchGet *dataloader.Loader
}

func (r *NotificationLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *NotificationLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *NotificationLoader) Get(
	ctx context.Context,
	id string,
) (*data.Notification, error) {
	notificationData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	notification, ok := notificationData.(*data.Notification)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return notification, nil
}

func (r *NotificationLoader) GetMany(
	ctx context.Context,
	ids *[]string,
) ([]*data.Notification, []error) {
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	notificationData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("errors", errs).Error(util.Trace(""))
		return nil, errs
	}
	notifications := make([]*data.Notification, len(notificationData))
	for i, d := range notificationData {
		var ok bool
		notifications[i], ok = d.(*data.Notification)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return notifications, nil
}
