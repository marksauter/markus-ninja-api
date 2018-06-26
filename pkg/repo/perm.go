package repo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fatih/structs"
	"github.com/iancoleman/strcase"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

func NewPermRepo(svc *data.PermService) *PermRepo {
	return &PermRepo{svc: svc}
}

type PermRepo struct {
	load   *loader.QueryPermLoader
	svc    *data.PermService
	viewer *data.User
}

func (r *PermRepo) Open(ctx context.Context) error {
	if r.load == nil {
		var ok bool
		r.viewer, ok = myctx.UserFromContext(ctx)
		if !ok {
			return fmt.Errorf("viewer not found")
		}
		r.load = loader.NewQueryPermLoader(r.svc, r.viewer)
	}
	return nil
}

func (r *PermRepo) Close() {
	r.load = nil
}

func (r *PermRepo) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("permission connection closed")
		return ErrConnClosed
	}
	return nil
}

func (r *PermRepo) Clear(o perm.Operation) {
	r.load.Clear(o)
}

func (r *PermRepo) ClearAll() {
	r.load.ClearAll()
}

func (r *PermRepo) Check(a perm.AccessLevel, node interface{}) (FieldPermissionFunc, error) {
	var checkField FieldPermissionFunc
	if err := r.CheckConnection(); err != nil {
		return checkField, err
	}
	nt, err := perm.ParseNodeType(structs.Name(node))
	if err != nil {
		return checkField, err
	}
	o := perm.Operation{a, nt}

	additionalRoles := []string{}
	if a != perm.Create {
		if ok := r.viewerCanAdmin(node); ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	}
	queryPerm, err := r.load.Get(o, additionalRoles)
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
	if a == perm.Create || a == perm.Update {
		for _, f := range structs.Fields(node) {
			if !f.IsZero() {
				if ok := checkField(strcase.ToSnake(f.Name())); !ok {
					return checkField, ErrAccessDenied
				}
			}
		}
	}
	return checkField, nil
}

func (r *PermRepo) viewerCanAdmin(node interface{}) bool {
	vid := r.viewer.Id.String
	switch node := node.(type) {
	case data.Email:
		return vid == node.UserId.String
	case *data.Email:
		return vid == node.UserId.String
	case data.EVT:
		return vid == node.UserId.String
	case *data.EVT:
		return vid == node.UserId.String
	case data.Lesson:
		return vid == node.UserId.String
	case *data.Lesson:
		return vid == node.UserId.String
	case data.LessonComment:
		return vid == node.UserId.String
	case *data.LessonComment:
		return vid == node.UserId.String
	case data.Notification:
		return vid == node.UserId.String
	case *data.Notification:
		return vid == node.UserId.String
	case data.PRT:
		return vid == node.UserId.String
	case *data.PRT:
		return vid == node.UserId.String
	case data.Study:
		return vid == node.UserId.String
	case *data.Study:
		return vid == node.UserId.String
	case data.User:
		return vid == node.Id.String
	case *data.User:
		return vid == node.Id.String
	case data.UserAsset:
		return vid == node.UserId.String
	case *data.UserAsset:
		return vid == node.UserId.String
	default:
		return false
	}
}

// Middleware
func (r *PermRepo) Use(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.Open(req.Context())
		defer r.Close()
		h.ServeHTTP(rw, req)
	})
}
