package repo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

func NewPermRepo(svc *data.PermService) *PermRepo {
	return &PermRepo{svc: svc}
}

type PermRepo struct {
	load *loader.QueryPermLoader
	svc  *data.PermService
}

func (r *PermRepo) Open(ctx context.Context) error {
	if r.load == nil {
		viewer, ok := data.UserFromContext(ctx)
		if !ok {
			return fmt.Errorf("viewer not found")
		}
		r.load = loader.NewQueryPermLoader(r.svc, viewer.Roles...)
	}
	return nil
}

func (r *PermRepo) Close() {
	r.load = nil
}

func (r *PermRepo) Clear(o perm.Operation) {
	r.load.Clear(o)
}

func (r *PermRepo) ClearAll() {
	r.load.ClearAll()
}

func (r *PermRepo) Check(o perm.Operation) (FieldPermissionFunc, error) {
	var checkField FieldPermissionFunc
	if r.load == nil {
		mylog.Log.Error("permission connection closed")
		return checkField, ErrConnClosed
	}
	queryPerm, err := r.load.Get(o)
	if err != nil {
		if err == data.ErrNotFound {
			return checkField, ErrAccessDenied
		} else {
			return checkField, err
		}
	}
	checkField = func(field string) bool {
		for _, f := range queryPerm.Fields {
			if f == field {
				return true
			}
		}
		return false
	}
	return checkField, nil
}

// Middleware
func (r *PermRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
