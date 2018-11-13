package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/loader"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

const Guest = "guest"

func NewPermitter(repos *Repos, conf *myconf.Config) *Permitter {
	return &Permitter{
		load:  loader.NewQueryPermLoader(),
		repos: repos,
	}
}

type Permitter struct {
	conf  *myconf.Config
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
	if node == nil {
		err := errors.New("node is nil")
		mylog.Log.WithError(err).Error(util.Trace(""))
		return f, err
	} else if !structs.IsStruct(node) {
		err := errors.New("node is not a struct")
		mylog.Log.WithField("node", node).WithError(err).Error(util.Trace(""))
		return f, err
	}

	nt, err := mytype.ParseNodeType(structs.Name(node))
	if err != nil {
		return f, err
	}
	o := mytype.NewOperation(a, nt)

	// If we are attempting to read the object, then check if the viewer has
	// access to the object.
	if a == mytype.ReadAccess {
		ok, err := r.ViewerCanRead(ctx, node)
		if err != nil {
			return f, err
		} else if !ok {
			return f, ErrAccessDenied
		}
	}

	additionalRoles := []string{}
	// If we are not creating or connecting, then check if the viewer can admin
	// the object. If yes, then grant the owner role to the user.
	if a != mytype.CreateAccess && a != mytype.ConnectAccess {
		ok, err := r.ViewerCanAdmin(ctx, node)
		if err != nil {
			return f, err
		}
		if ok {
			additionalRoles = append(additionalRoles, data.OwnerRole)
		}
	} else {
		// If we are creating/connecting, then check if viewer can create the object
		// with the Owner role. If yes, then grant the Owner role to the user.
		ok, err := r.ViewerCanCreateWithOwnership(ctx, node)
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
					return f, ErrFieldAccessDenied
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
					return f, ErrFieldAccessDenied
				}
			}
		}
	}
	return f, nil
}

// Can the viewer read the node?
func (r *Permitter) ViewerCanRead(
	ctx context.Context,
	node interface{},
) (bool, error) {
	switch node := node.(type) {
	case data.Lesson:
		// If the lesson has not been published, then check if the viewer can admin
		// the object
		if node.PublishedAt.Status == pgtype.Undefined || node.PublishedAt.Status == pgtype.Null {
			return r.ViewerCanAdmin(ctx, node)
		}
	case *data.Lesson:
		// If the lesson has not been published, then check if the viewer can admin
		// the object
		if node.PublishedAt.Status == pgtype.Undefined || node.PublishedAt.Status == pgtype.Null {
			return r.ViewerCanAdmin(ctx, node)
		}
	case data.LessonComment:
		// If the lesson comment has not been published, then check if the viewer can admin
		// the object
		if node.PublishedAt.Status == pgtype.Undefined || node.PublishedAt.Status == pgtype.Null {
			return r.ViewerCanAdmin(ctx, node)
		}
	case *data.LessonComment:
		// If the lesson comment has not been published, then check if the viewer can admin
		// the object
		if node.PublishedAt.Status == pgtype.Undefined || node.PublishedAt.Status == pgtype.Null {
			return r.ViewerCanAdmin(ctx, node)
		}
	default:
		return true, nil
	}
	return true, nil
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
	if viewer.Login.String == Guest {
		return false, nil
	}
	vid := viewer.ID.String
	switch node := node.(type) {
	case data.Appled:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			appled, err := r.repos.Appled().load.Get(ctx, node.ID.Int)
			if err != nil {
				return false, err
			}
			userID = &appled.UserID
		}
		return vid == userID.String, nil
	case *data.Appled:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			appled, err := r.repos.Appled().load.Get(ctx, node.ID.Int)
			if err != nil {
				return false, err
			}
			userID = &appled.UserID
		}
		return vid == userID.String, nil
	case data.Course:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			course, err := r.repos.Course().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &course.UserID
		}
		return vid == userID.String, nil
	case *data.Course:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			course, err := r.repos.Course().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &course.UserID
		}
		return vid == userID.String, nil
	case data.CourseLesson:
		courseID := &node.CourseID
		if courseID.Status == pgtype.Undefined {
			courseLesson, err := r.repos.CourseLesson().load.Get(ctx, node.LessonID.String)
			if err != nil {
				return false, err
			}
			courseID = &courseLesson.CourseID
		}
		course, err := r.repos.Course().load.Get(ctx, courseID.String)
		if err != nil {
			return false, err
		}
		return vid == course.UserID.String, nil
	case *data.CourseLesson:
		courseID := &node.CourseID
		if courseID.Status == pgtype.Undefined {
			courseLesson, err := r.repos.CourseLesson().load.Get(ctx, node.LessonID.String)
			if err != nil {
				return false, err
			}
			courseID = &courseLesson.CourseID
		}
		course, err := r.repos.Course().load.Get(ctx, courseID.String)
		if err != nil {
			return false, err
		}
		return vid == course.UserID.String, nil
	case data.Email:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			email, err := r.repos.Email().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &email.UserID
		}
		return vid == userID.String, nil
	case *data.Email:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			email, err := r.repos.Email().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &email.UserID
		}
		return vid == userID.String, nil
	case data.Enrolled:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			enrolled, err := r.repos.Enrolled().load.Get(ctx, node.ID.Int)
			if err != nil {
				return false, err
			}
			userID = &enrolled.UserID
		}
		return vid == userID.String, nil
	case *data.Enrolled:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			enrolled, err := r.repos.Enrolled().load.Get(ctx, node.ID.Int)
			if err != nil {
				return false, err
			}
			userID = &enrolled.UserID
		}
		return vid == userID.String, nil
	case data.EVT:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			evt, err := r.repos.EVT().load.Get(ctx, node.EmailID.String, node.Token.String)
			if err != nil {
				return false, err
			}
			userID = &evt.UserID
		}
		return vid == userID.String, nil
	case *data.EVT:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			evt, err := r.repos.EVT().load.Get(ctx, node.EmailID.String, node.Token.String)
			if err != nil {
				return false, err
			}
			userID = &evt.UserID
		}
		return vid == userID.String, nil
	case data.Label:
		label, err := r.repos.Label().load.Get(ctx, node.ID.String)
		if err != nil {
			return false, err
		}
		study, err := r.repos.Study().load.Get(ctx, label.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case *data.Label:
		label, err := r.repos.Label().load.Get(ctx, node.ID.String)
		if err != nil {
			return false, err
		}
		study, err := r.repos.Study().load.Get(ctx, label.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case data.Labeled:
		userID := mytype.OID{}
		switch node.LabelableID.Type {
		case "Lesson":
			lesson, err := r.repos.Lesson().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lesson.UserID
		case "LessonComment":
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lessonComment.UserID
		}
		return vid == userID.String, nil
	case *data.Labeled:
		userID := mytype.OID{}
		switch node.LabelableID.Type {
		case "Lesson":
			lesson, err := r.repos.Lesson().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lesson.UserID
		case "LessonComment":
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lessonComment.UserID
		}
		return vid == userID.String, nil
	case data.Lesson:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			lesson, err := r.repos.Lesson().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &lesson.UserID
		}
		return vid == userID.String, nil
	case *data.Lesson:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			lesson, err := r.repos.Lesson().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &lesson.UserID
		}
		return vid == userID.String, nil
	case data.LessonComment:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &lessonComment.UserID
		}
		return vid == userID.String, nil
	case *data.LessonComment:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &lessonComment.UserID
		}
		return vid == userID.String, nil
	case data.Notification:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			notification, err := r.repos.Notification().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &notification.UserID
		}
		return vid == userID.String, nil
	case *data.Notification:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			notification, err := r.repos.Notification().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &notification.UserID
		}
		return vid == userID.String, nil
	case data.PRT:
		return vid == node.UserID.String, nil
	case *data.PRT:
		return vid == node.UserID.String, nil
	case data.Study:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			study, err := r.repos.Study().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &study.UserID
		}
		return vid == userID.String, nil
	case *data.Study:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			study, err := r.repos.Study().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &study.UserID
		}
		return vid == userID.String, nil
	case data.Topiced:
		userID := mytype.OID{}
		switch node.TopicableID.Type {
		case "Course":
			course, err := r.repos.Course().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = course.UserID
		case "Study":
			study, err := r.repos.Study().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = study.UserID
		}
		return vid == userID.String, nil
	case *data.Topiced:
		userID := mytype.OID{}
		switch node.TopicableID.Type {
		case "Course":
			course, err := r.repos.Course().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = course.UserID
		case "Study":
			study, err := r.repos.Study().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = study.UserID
		}
		return vid == userID.String, nil
	case data.User:
		return vid == node.ID.String, nil
	case *data.User:
		return vid == node.ID.String, nil
	case data.UserAsset:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			userAsset, err := r.repos.UserAsset().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &userAsset.UserID
		}
		return vid == userID.String, nil
	case *data.UserAsset:
		userID := &node.UserID
		if node.UserID.Status == pgtype.Undefined {
			userAsset, err := r.repos.UserAsset().load.Get(ctx, node.ID.String)
			if err != nil {
				return false, err
			}
			userID = &userAsset.UserID
		}
		return vid == userID.String, nil
	default:
		return false, nil
	}
	return false, nil
}

// Can the viewer create the passed node with the Owner role?
// Mainly used for objects that have parent objects, and the viewer must
// be the owner of the parent object to create a child object.
func (r *Permitter) ViewerCanCreateWithOwnership(
	ctx context.Context,
	node interface{},
) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, &myctx.ErrNotFound{"viewer"}
	}
	if viewer.Login.String == Guest {
		return false, nil
	}
	vid := viewer.ID.String
	switch node := node.(type) {
	case data.Course:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case *data.Course:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case data.CourseLesson:
		course, err := r.repos.Course().load.Get(ctx, node.CourseID.String)
		if err != nil {
			return false, err
		}
		return vid == course.UserID.String, nil
	case *data.CourseLesson:
		course, err := r.repos.Course().load.Get(ctx, node.CourseID.String)
		if err != nil {
			return false, err
		}
		return vid == course.UserID.String, nil
	case data.Label:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case *data.Label:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case data.Labeled:
		userID := mytype.OID{}
		switch node.LabelableID.Type {
		case "Lesson":
			lesson, err := r.repos.Lesson().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lesson.UserID
		case "LessonComment":
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lessonComment.UserID
		}
		return vid == userID.String, nil
	case *data.Labeled:
		userID := mytype.OID{}
		switch node.LabelableID.Type {
		case "Lesson":
			lesson, err := r.repos.Lesson().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lesson.UserID
		case "LessonComment":
			lessonComment, err := r.repos.LessonComment().load.Get(ctx, node.LabelableID.String)
			if err != nil {
				return false, err
			}
			userID = lessonComment.UserID
		}
		return vid == userID.String, nil
	case data.Lesson:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case *data.Lesson:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case data.Topiced:
		userID := mytype.OID{}
		switch node.TopicableID.Type {
		case "Course":
			course, err := r.repos.Course().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = course.UserID
		case "Study":
			study, err := r.repos.Study().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = study.UserID
		}
		return vid == userID.String, nil
	case *data.Topiced:
		userID := mytype.OID{}
		switch node.TopicableID.Type {
		case "Course":
			course, err := r.repos.Course().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = course.UserID
		case "Study":
			study, err := r.repos.Study().load.Get(ctx, node.TopicableID.String)
			if err != nil {
				return false, err
			}
			userID = study.UserID
		}
		return vid == userID.String, nil
	case data.UserAsset:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	case *data.UserAsset:
		study, err := r.repos.Study().load.Get(ctx, node.StudyID.String)
		if err != nil {
			return false, err
		}
		return vid == study.UserID.String, nil
	default:
		return false, nil
	}
	return false, nil
}
