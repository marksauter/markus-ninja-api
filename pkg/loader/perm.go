package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

func NewQueryPermLoader(svc *data.PermissionService, viewer *data.User) *QueryPermLoader {
	loader := &QueryPermLoader{
		svc:    svc,
		viewer: viewer,
	}
	loader.batchGet = createLoader(
		func(
			ctx context.Context,
			keys dataloader.Keys,
		) []*dataloader.Result {
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
					queryPerm, err := svc.GetQueryPermission(operation, ks[1:])
					results[i] = &dataloader.Result{Data: queryPerm, Error: err}
				}(i, key)
			}

			wg.Wait()

			return results
		},
	)
	return loader
}

type QueryPermLoader struct {
	svc    *data.PermissionService
	viewer *data.User

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
	o *mytype.Operation,
	roles []string,
) (*data.QueryPermission, error) {
	ctx := context.Background()
	for _, role := range l.viewer.Roles.Elements {
		roles = append(roles, role.String)
	}
	keyParts := append([]string{o.String()}, roles...)
	compositeKey := newCompositeKey(keyParts...)
	permData, err := l.batchGet.Load(ctx, compositeKey)()
	if err != nil {
		return nil, err
	}
	perm, ok := permData.(*data.QueryPermission)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return perm, nil
}

func (l *QueryPermLoader) GetMany(os []mytype.Operation) ([]*data.QueryPermission, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(os))
	for i, o := range os {
		keys[i] = dataloader.StringKey(o.String())
	}
	permData, errs := l.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	perms := make([]*data.QueryPermission, len(permData))
	for i, d := range permData {
		var ok bool
		perms[i], ok = d.(*data.QueryPermission)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return perms, nil
}

func (l *QueryPermLoader) batchGetFunc(
	ctx context.Context,
	keys dataloader.Keys,
) []*dataloader.Result {
	var (
		n       = len(keys)
		results = make([]*dataloader.Result, n)
		wg      sync.WaitGroup
	)

	wg.Add(n)

	for i, key := range keys {
		go func(i int, key dataloader.Key) {
			defer wg.Done()
			operation, err := mytype.ParseOperation(key.String())
			if err != nil {
				results[i] = &dataloader.Result{Data: nil, Error: err}
				return
			}
			var roles []string
			err = l.viewer.Roles.AssignTo(roles)
			if err != nil {
				results[i] = &dataloader.Result{Data: nil, Error: err}
				return
			}
			queryPerm, err := l.svc.GetQueryPermission(operation, roles)
			results[i] = &dataloader.Result{Data: queryPerm, Error: err}
		}(i, key)
	}

	wg.Wait()

	return results
}
