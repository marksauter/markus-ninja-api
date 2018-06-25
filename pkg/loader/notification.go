package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
)

func NewNotificationLoader(svc *data.NotificationService) *NotificationLoader {
	return &NotificationLoader{
		svc:      svc,
		batchGet: createLoader(newBatchGetNotificationFn(svc.Get)),
	}
}

type NotificationLoader struct {
	svc *data.NotificationService

	batchGet *dataloader.Loader
}

func (r *NotificationLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *NotificationLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *NotificationLoader) Get(id string) (*data.Notification, error) {
	ctx := context.Background()
	notificationData, err := r.batchGet.Load(ctx, dataloader.StringKey(id))()
	if err != nil {
		return nil, err
	}
	notification, ok := notificationData.(*data.Notification)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return notification, nil
}

func (r *NotificationLoader) GetMany(ids *[]string) ([]*data.Notification, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*ids))
	for i, k := range *ids {
		keys[i] = dataloader.StringKey(k)
	}
	notificationData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	notifications := make([]*data.Notification, len(notificationData))
	for i, d := range notificationData {
		var ok bool
		notifications[i], ok = d.(*data.Notification)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return notifications, nil
}

func newBatchGetNotificationFn(
	getter func(string) (*data.Notification, error),
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
				notification, err := getter(key.String())
				results[i] = &dataloader.Result{Data: notification, Error: err}
			}(i, key)
		}

		wg.Wait()

		return results
	}
}
