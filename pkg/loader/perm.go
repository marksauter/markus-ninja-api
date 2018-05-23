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

func (l *QueryPermLoader) Clear(o perm.Operation) {
	ctx := context.Background()
	l.batchGet.Clear(ctx, dataloader.StringKey(o.String()))
}

func (l *QueryPermLoader) ClearAll() {
	l.batchGet.ClearAll()
}

func (l *QueryPermLoader) AddRoles(roles ...data.Role) *QueryPermLoader {
	for _, r := range roles {
		l.viewerRoles = append(l.viewerRoles, r.String())
	}
	return l
}

func (l *QueryPermLoader) RemoveRoles(roles ...data.Role) *QueryPermLoader {
	viewerRoles := make([]string, 0, len(l.viewerRoles))
	for _, vr := range l.viewerRoles {
		remove := false
		for _, r := range roles {
			if vr == r.String() {
				remove = true
			}
		}
		if !remove {
			viewerRoles = append(viewerRoles, vr)
		}
		remove = false
	}
	l.viewerRoles = viewerRoles
	return l
}

func (l *QueryPermLoader) Get(o perm.Operation) (*perm.QueryPermission, error) {
	ctx := context.Background()
	permData, err := l.batchGet.Load(ctx, dataloader.StringKey(o.String()))()
	if err != nil {
		return nil, err
	}
	perm, ok := permData.(*perm.QueryPermission)
	if !ok {
		return nil, fmt.Errorf("wrong type")
	}

	return perm, nil
}

func (l *QueryPermLoader) GetMany(os []perm.Operation) ([]*perm.QueryPermission, []error) {
	ctx := context.Background()
	keys := make(dataloader.Keys, len(os))
	for i, o := range os {
		keys[i] = dataloader.StringKey(o.String())
	}
	permData, errs := l.batchGet.LoadMany(ctx, keys)()
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
			operation, err := perm.ParseOperation(key.String())
			if err != nil {
				results[i] = &dataloader.Result{Data: nil, Error: err}
				return
			}
			queryPerm, err := l.svc.GetQueryPermission(operation, l.viewerRoles...)
			results[i] = &dataloader.Result{Data: queryPerm, Error: err}
		}(i, key)
	}

	wg.Wait()

	return results
}
