package repo

import (
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/marksauter/markus-ninja-api/pkg/svccxn"
)

func NewPermRepo(svc *service.PermService) *PermRepo {
	return &PermRepo{svc: svc}
}

type PermRepo struct {
	cxn   *svccxn.PermConnection
	svc   *service.PermService
	perms map[string][]string
}

func (r *PermRepo) Open() {
	r.cxn = svccxn.NewPermConnection(r.svc)
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

func (r *PermRepo) checkConnection() bool {
	return r.cxn != nil
}

// Service methods

func (r *PermRepo) GetQueryPermission(
	o perm.Operation,
	roles ...string,
) (*perm.QueryPermission, error) {
	if ok := r.checkConnection(); !ok {
		mylog.Log.Error("perm connection closed")
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
