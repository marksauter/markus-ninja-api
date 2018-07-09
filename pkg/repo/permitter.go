package repo

import (
	"context"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
)

func NewPermitter(repos *Repos) *Permitter {
	return &Permitter{
		load:  loader.NewQueryPermLoader(),
		repos: repos,
	}
}

type Permitter struct {
	load  *loader.QueryPermLoader
	repos *Repos
}

func (r *Permitter) CheckConnection() error {
	if r.load == nil {
		mylog.Log.Error("permission connection closed")
		return ErrConnClosed
	}
	return nil
}

func (r *Permitter) ClearCacheOperation(o mytype.Operation) {
	r.load.Clear(o)
}

func (r *Permitter) ClearCache() {
	r.load.ClearAll()
}

func (r *Permitter) Check(
	ctx context.Context,
	a mytype.AccessLevel,
	node interface{},
) (f FieldPermissionFunc, err error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		err = myctx.ErrNotFound{"viewer"}
		return
	}
	if err = r.CheckConnection(); err != nil {
		return
	}
	nt, err := mytype.ParseNodeType(structs.Name(node))
	if err != nil {
		return
	}
	o := mytype.NewOperation(a, nt)

	additionalRoles := []string{}
	// If we are not creating, then check if the viewer can admin the object. If
	// yes, then grant the owner role to the user.
	if a != mytype.CreateAccess {
		ok, err := r.ViewerCanAdmin(viewer, node)
		if err != nil {
			return
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	} else {
		// If we are creating, then check if viewer can create the object.  If yes,
		// then grant the owner role to the user.
		ok, err := r.ViewerCanCreate(viewer, node)
		if err != nil {
			return
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	}
	// Get the query permissions.
	queryPerm, err := r.load.Get(o, additionalRoles)
	if err != nil {
		if err == data.ErrNotFound {
			err = ErrAccessDenied
			return
		} else {
			return
		}
	}
	// Set field permission function for the fields returned by the query
	// permission.
	f = func(field string) bool {
		// If the returned query permission has a null value for fields, then return
		// true for all checks.
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
		for _, field := range structs.Fields(node) {
			if !field.IsZero() {
				dbField := field.Tag("db")
				if ok := f(dbField); !ok {
					err = ErrAccessDenied
					return
				}
			}
		}
	}
	return
}

// Can the viewer admin the node, i.e. is the viewer the owner of the object?
func (r *Permitter) ViewerCanAdmin(viewer *data.User, node interface{}) (bool, error) {
	vid := viewer.Id.String
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
func (r *Permitter) ViewerCanCreate(viewer *data.User, node interface{}) (bool, error) {
	vid := viewer.Id.String
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
