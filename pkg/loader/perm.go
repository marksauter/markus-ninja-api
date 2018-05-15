package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

func NewQueryPermLoader(svc *data.PermService, viewerRoles ...string) *QueryPermLoader {
	loader := &QueryPermLoader{
		viewerRoles: viewerRoles,
		svc:         svc,
	}
	loader.batchGet = createLoader(loader.batchGetFunc)
	return loader
}

type QueryPermLoader struct {
	svc         *data.PermService
	viewerRoles []string

	batchGet *dataloader.Loader
}

func (r *QueryPermLoader) Clear(id string) {
	ctx := context.Background()
	r.batchGet.Clear(ctx, dataloader.StringKey(id))
}

func (r *QueryPermLoader) ClearAll() {
	r.batchGet.ClearAll()
}

func (r *QueryPermLoader) AddRoles(roles ...string) *QueryPermLoader {
	r.viewerRoles = append(r.viewerRoles, roles...)
	return r
}

func (r *QueryPermLoader) Get(operation string) (*perm.QueryPermission, error) {
	ctx := context.Background()
	permData, err := r.batchGet.Load(ctx, dataloader.StringKey(operation))()
	if err != nil {
		return nil, err
	}
	perm, ok := permData.(*perm.QueryPermission)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return perm, nil
}

func (r *QueryPermLoader) GetMany(operations *[]string) ([]*perm.QueryPermission, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(*operations))
	for i, o := range *operations {
		keys[i] = dataloader.StringKey(o)
	}
	permData, errs := r.batchGet.LoadMany(ctx, keys)()
	if errs != nil {
		return nil, errs
	}
	perms := make([]*perm.QueryPermission, len(permData))
	for i, d := range permData {
		var ok bool
		perms[i], ok = d.(*perm.QueryPermission)
		if !ok {
			return nil, []error{fmt.Errorf("wrong type")}
		}
	}

	return perms, nil
}

func (r *QueryPermLoader) batchGetFunc(
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
			operation, err := perm.ParseOperation(key.String())
			if err != nil {
				results[i] = &dataloader.Result{Data: nil, Error: err}
				return
			}
			queryPerm, err := r.svc.GetQueryPermission(operation, r.viewerRoles...)
			results[i] = &dataloader.Result{Data: queryPerm, Error: err}
		}(i, key)
	}

	wg.Wait()

	return results
}
