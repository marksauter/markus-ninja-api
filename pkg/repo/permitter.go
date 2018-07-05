package repo

import (
	"context"
	"fmt"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
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

func (r *Permitter) Clear(o mytype.Operation) {
	r.load.Clear(o)
}

func (r *Permitter) ClearAll() {
	r.load.ClearAll()
}

func (r *Permitter) Check(a mytype.AccessLevel, node interface{}) (FieldPermissionFunc, error) {
	var checkField FieldPermissionFunc
	if err := r.CheckConnection(); err != nil {
		return checkField, err
	}
	nt, err := mytype.ParseNodeType(structs.Name(node))
	if err != nil {
		return checkField, err
	}
	o := mytype.NewOperation(a, nt)

	additionalRoles := []string{}
	// If we are not creating, then check if the viewer can admin the object. If
	// yes, then grant the owner role to the user.
	if a != mytype.CreateAccess {
		ok, err := r.ViewerCanAdmin(node)
		if err != nil {
			return checkField, err
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	} else {
		// If we are creating, then check if viewer can create the object.  If yes,
		// then grant the owner role to the user.
		ok, err := r.ViewerCanCreate(node)
		if err != nil {
			return checkField, err
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	}
	// Get the query permissions.
	queryPerm, err := r.load.Get(o, additionalRoles)
	if err != nil {
		if err == data.ErrNotFound {
			return checkField, ErrAccessDenied
		} else {
			return checkField, err
		}
	}
	// Set field permission function for the fields returned by the query
	// permission.
	checkField = func(field string) bool {
		// If the returned query permission has a null value for fields, then return
		// true for all fields.
		// NOTE: checkField only makes sense in respect to create/read/update
		// operations.
		if queryPerm.Fields.Status == pgtype.Null {
			return true
		}
		for _, f := range queryPerm.Fields.Elements {
			if f.String == field {
				return true
			}
		}
		return false
	}
	// If creating/updating, then check if fields provided are permitted.
	if a == mytype.CreateAccess || a == mytype.UpdateAccess {
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

// Can the viewer admin the node, i.e. is the viewer the owner of the object?
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

// Can the viewer create the passed node? Mainly used for objects that have
// parent objects, and the viewer must be the owner of the parent object to
// create a child object.
func (r *Permitter) ViewerCanCreate(node interface{}) (bool, error) {
	vid := r.viewer.Id.String
	switch node := node.(type) {
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
		study, err := r.repos.Study().Get(node.StudyId.String)
		if err != nil {
			return false, err
		}
		userId, err := study.UserId()
		if err != nil {
			return false, err
		}
		return vid == userId.String, nil
	case *data.Lesson:
		study, err := r.repos.Study().Get(node.StudyId.String)
		if err != nil {
			return false, err
		}
		userId, err := study.UserId()
		if err != nil {
			return false, err
		}
		return vid == userId.String, nil
	default:
		return false, nil
	}
	return false, nil
}
