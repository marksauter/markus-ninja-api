package repo

import (
	"context"
	"fmt"

	"github.com/fatih/structs"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
)

func NewPermitter(svc *data.PermissionService, repos *Repos) *Permitter {
	return &Permitter{
		repos: repos,
		svc:   svc,
	}
}

type Permitter struct {
	load   *loader.QueryPermLoader
	repos  *Repos
	svc    *data.PermissionService
	viewer *data.User
}

func (r *Permitter) Begin(ctx context.Context) error {
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

func (r *Permitter) End() {
	r.load.ClearAll()
}

func (r *Permitter) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("permission connection closed")
		return ErrConnClosed
	}
	return nil
}

func (r *Permitter) Clear(o perm.Operation) {
	r.load.Clear(o)
}

func (r *Permitter) ClearAll() {
	r.load.ClearAll()
}

func (r *Permitter) Check(a perm.AccessLevel, node interface{}) (FieldPermissionFunc, error) {
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
		ok, err := r.ViewerCanAdmin(node)
		if err != nil {
			return checkField, err
		}
		if ok {
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
				dbField := f.Tag("db")
				if ok := checkField(dbField); !ok {
					return checkField, ErrAccessDenied
				}
			}
		}
	}
	return checkField, nil
}

func (r *Permitter) ViewerCanAdmin(node interface{}) (bool, error) {
	vid := r.viewer.Id.String
	switch node := node.(type) {
	case data.Email:
		return vid == node.UserId.String, nil
	case *data.Email:
		return vid == node.UserId.String, nil
	case data.EVT:
		return vid == node.UserId.String, nil
	case *data.EVT:
		return vid == node.UserId.String, nil
	case data.Label:
		study, err := r.repos.Study().Get(node.StudyId.String)
		if err != nil {
			return false, err
		}
		userId, err := study.UserId()
		if err != nil {
			return false, err
		}
		return vid == userId.String, nil
	case *data.Label:
		study, err := r.repos.Study().Get(node.StudyId.String)
		if err != nil {
			return false, err
		}
		userId, err := study.UserId()
		if err != nil {
			return false, err
		}
		return vid == userId.String, nil
	case data.Lesson:
		return vid == node.UserId.String, nil
	case *data.Lesson:
		return vid == node.UserId.String, nil
	case data.LessonComment:
		return vid == node.UserId.String, nil
	case *data.LessonComment:
		return vid == node.UserId.String, nil
	case data.Notification:
		return vid == node.UserId.String, nil
	case *data.Notification:
		return vid == node.UserId.String, nil
	case data.PRT:
		return vid == node.UserId.String, nil
	case *data.PRT:
		return vid == node.UserId.String, nil
	case data.Study:
		return vid == node.UserId.String, nil
	case *data.Study:
		return vid == node.UserId.String, nil
	case data.User:
		return vid == node.Id.String, nil
	case *data.User:
		return vid == node.Id.String, nil
	case data.UserAsset:
		return vid == node.UserId.String, nil
	case *data.UserAsset:
		return vid == node.UserId.String, nil
	default:
		return false, nil
	}
	return false, nil
}
