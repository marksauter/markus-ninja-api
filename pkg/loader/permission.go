package loader

import (
	"context"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

func NewQueryPermLoader() *QueryPermLoader {
	return &QueryPermLoader{
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
						ks := splitCompositeKey(key)
						operation, err := mytype.ParseOperation(ks[0])
						if err != nil {
							results[i] = &dataloader.Result{Data: nil, Error: err}
							return
						}
						db, ok := myctx.QueryerFromContext(ctx)
						if !ok {
							results[i] = &dataloader.Result{Error: &myctx.ErrNotFound{"queryer"}}
							return
						}
						queryPerm, err := data.GetQueryPermission(db, operation, ks[1:])
						results[i] = &dataloader.Result{Data: queryPerm, Error: err}
					}(i, key)
				}

				wg.Wait()

				return results
			},
		),
	}
}

type QueryPermLoader struct {
	batchGet *dataloader.Loader
}

func (l *QueryPermLoader) Clear(o mytype.Operation) {
	ctx := context.Background()
	l.batchGet.Clear(ctx, dataloader.StringKey(o.String()))
}

func (l *QueryPermLoader) ClearAll() {
	l.batchGet.ClearAll()
}

func (l *QueryPermLoader) Get(
	ctx context.Context,
	o *mytype.Operation,
	roles []string,
) (*data.QueryPermission, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, &myctx.ErrNotFound{"viewer"}
	}
	for _, role := range viewer.Roles.Elements {
		roles = append(roles, role.String)
	}
	keyParts := append([]string{o.String()}, roles...)
	compositeKey := newCompositeKey(keyParts...)
	permData, err := l.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	perm, ok := permData.(*data.QueryPermission)
	if !ok {
		err := ErrWrongType
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}

	return perm, nil
}

func (l *QueryPermLoader) GetMany(
	ctx context.Context,
	os []mytype.Operation,
) ([]*data.QueryPermission, []error) {
	keys := make(dataloader.Keys, len(os))
	for i, o := range os {
		keys[i] = dataloader.StringKey(o.String())
	}
	permData, errs := l.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		mylog.Log.WithField("erros", errs).Error(util.Trace(""))
		return nil, errs
	}
	perms := make([]*data.QueryPermission, len(permData))
	for i, d := range permData {
		var ok bool
		perms[i], ok = d.(*data.QueryPermission)
		if !ok {
			err := ErrWrongType
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, []error{err}
		}
	}

	return perms, nil
}
