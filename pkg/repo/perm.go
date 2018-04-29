package repo

import (
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
	cxn   *loader.PermLoader
	svc   *data.PermService
	perms map[string][]string
}

func (r *PermRepo) Open() {
	r.cxn = loader.NewPermLoader(r.svc)
}

func (r *PermRepo) Close() {
	r.cxn = nil
	r.perms = nil
}

func (r *PermRepo) AddPermission(p perm.QueryPermission) {
	r.perms[p.Operation.String()] = p.Fields
}

func (r *PermRepo) CheckPermission(o perm.Operation) (func(string) bool, bool) {
	fields, ok := r.perms[o.String()]
	checkField := func(field string) bool {
		for _, f := range fields {
			if f == field {
				return true
			}
		}
		return false
	}
	return checkField, ok
}

func (r *PermRepo) ClearPermissions() {
	r.perms = nil
}

func (r *PermRepo) checkLoader() bool {
	return r.cxn != nil
}

// Service methods

func (r *PermRepo) GetQueryPermission(
	o perm.Operation,
	roles ...string,
) (*perm.QueryPermission, error) {
	if ok := r.checkLoader(); !ok {
		mylog.Log.Error("permission connection closed")
		return nil, ErrConnClosed
	}
	return r.cxn.GetQueryPermission(o, roles...)
}

// Middleware
func (r *PermRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open()
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
