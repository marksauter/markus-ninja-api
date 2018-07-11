package repo

import (
	"context"
	"strings"

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
) (FieldPermissionFunc, error) {
	var f FieldPermissionFunc
	if err := r.CheckConnection(); err != nil {
		return f, err
	}
	nt, err := mytype.ParseNodeType(structs.Name(node))
	if err != nil {
		return f, err
	}
	o := mytype.NewOperation(a, nt)

	additionalRoles := []string{}
	// If we are not creating, then check if the viewer can admin the object. If
	// yes, then grant the owner role to the user.
	if a != mytype.CreateAccess {
		ok, err := r.ViewerCanAdmin(ctx, node)
		if err != nil {
			return f, err
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	} else {
		// If we are creating, then check if viewer can create the object.  If yes,
		// then grant the owner role to the user.
		ok, err := r.ViewerCanCreate(ctx, node)
		if err != nil {
			return f, err
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	}
	// Get the query permissions.
	queryPerm, err := r.load.Get(ctx, o, additionalRoles)
	if err != nil {
		if err == data.ErrNotFound {
			return f, ErrAccessDenied
		} else {
			return f, err
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
	if a == mytype.CreateAccess {
		for _, field := range structs.Fields(node) {
			permit := field.Tag("permit")
			createable := strings.Contains(permit, "create")
			if createable && !field.IsZero() {
				dbField := field.Tag("db")
				if ok := f(dbField); !ok {
					return f, ErrAccessDenied
				}
			}
		}
	} else if a == mytype.UpdateAccess {
		for _, field := range structs.Fields(node) {
			permit := field.Tag("permit")
			updateable := strings.Contains(permit, "update")
			if updateable && !field.IsZero() {
				dbField := field.Tag("db")
				if ok := f(dbField); !ok {
					return f, ErrAccessDenied
				}
			}
		}
	}
	return f, nil
}

// Can the viewer admin the node, i.e. is the viewer the owner of the object?
func (r *Permitter) ViewerCanAdmin(
	ctx context.Context,
	node interface{},
) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, &myctx.ErrNotFound{"viewer"}
	}
	vid := viewer.Id.String
	switch node := node.(type) {
	case data.Email:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			email, err := r.repos.Email().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &email.UserId
		}
		return vid == userId.String, nil
	case *data.Email:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			email, err := r.repos.Email().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &email.UserId
		}
		return vid == userId.String, nil
	case data.EVT:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			evt, err := r.repos.EVT().load.Get(ctx, node.EmailId.String, node.Token.String)
			if err != nil {
				return false, err
			}
			userId = &evt.UserId
		}
		return vid == userId.String, nil
	case *data.EVT:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			evt, err := r.repos.EVT().load.Get(ctx, node.EmailId.String, node.Token.String)
			if err != nil {
				return false, err
			}
			userId = &evt.UserId
		}
		return vid == userId.String, nil
	case data.Label:
		label, err := r.repos.Label().load.Get(ctx, node.Id.String)
		if err != nil {
			return false, err
		}
		study, err := r.repos.Study().load.Get(ctx, label.StudyId.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserId.String, nil
	case *data.Label:
		label, err := r.repos.Label().load.Get(ctx, node.Id.String)
		if err != nil {
			return false, err
		}
		study, err := r.repos.Study().load.Get(ctx, label.StudyId.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserId.String, nil
	case data.Lesson:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			lesson, err := r.repos.Lesson().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &lesson.UserId
		}
		return vid == userId.String, nil
	case *data.Lesson:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			lesson, err := r.repos.Lesson().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &lesson.UserId
		}
		return vid == userId.String, nil
	case data.LessonComment:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &lessonComment.UserId
		}
		return vid == userId.String, nil
	case *data.LessonComment:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &lessonComment.UserId
		}
		return vid == userId.String, nil
	case data.Notification:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			notification, err := r.repos.Notification().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &notification.UserId
		}
		return vid == userId.String, nil
	case *data.Notification:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			notification, err := r.repos.Notification().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &notification.UserId
		}
		return vid == userId.String, nil
	case data.PRT:
		return vid == node.UserId.String, nil
	case *data.PRT:
		return vid == node.UserId.String, nil
	case data.Study:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			study, err := r.repos.Study().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &study.UserId
		}
		return vid == userId.String, nil
	case *data.Study:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			study, err := r.repos.Study().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &study.UserId
		}
		return vid == userId.String, nil
	case data.User:
		return vid == node.Id.String, nil
	case *data.User:
		return vid == node.Id.String, nil
	case data.UserAsset:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			userAsset, err := r.repos.UserAsset().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &userAsset.UserId
		}
		return vid == userId.String, nil
	case *data.UserAsset:
		userId := &node.UserId
		if node.UserId.Status == pgtype.Undefined {
			userAsset, err := r.repos.UserAsset().load.Get(ctx, node.Id.String)
			if err != nil {
				return false, err
			}
			userId = &userAsset.UserId
		}
		return vid == userId.String, nil
	default:
		return false, nil
	}
	return false, nil
}

// Can the viewer create the passed node? Mainly used for objects that have
// parent objects, and the viewer must be the owner of the parent object to
// create a child object.
func (r *Permitter) ViewerCanCreate(
	ctx context.Context,
	node interface{},
) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, &myctx.ErrNotFound{"viewer"}
	}
	vid := viewer.Id.String
	switch node := node.(type) {
	case data.Label:
		study, err := r.repos.Study().load.Get(ctx, node.StudyId.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserId.String, nil
	case *data.Label:
		study, err := r.repos.Study().load.Get(ctx, node.StudyId.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserId.String, nil
	case data.Lesson:
		study, err := r.repos.Study().load.Get(ctx, node.StudyId.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserId.String, nil
	case *data.Lesson:
		study, err := r.repos.Study().load.Get(ctx, node.StudyId.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserId.String, nil
	default:
		return false, nil
	}
	return false, nil
}
