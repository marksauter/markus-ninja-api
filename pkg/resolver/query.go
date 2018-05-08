package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func (r *RootResolver) Node(
	ctx context.Context,
	args struct{ Id string },
) (*nodeResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		mylog.Log.Error("viewer not found")
	}
	parsedId, err := oid.Parse(args.Id)
	if err != nil {
		return nil, err
	}
	switch parsedId.Type() {
	case "Lesson":
		queryPerm, err := r.Repos.Perm().GetQueryPermission(
			perm.ReadLesson,
			viewer.Roles()...,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error retrieving query permission")
			return nil, repo.ErrAccessDenied
		}
		r.Repos.Lesson().AddPermission(queryPerm)
		lesson, err := r.Repos.Lesson().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&lessonResolver{Lesson: lesson, Repos: r.Repos}}, nil
	case "Study":
		queryPerm, err := r.Repos.Perm().GetQueryPermission(
			perm.ReadStudy,
			viewer.Roles()...,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error retrieving query permission")
			return nil, repo.ErrAccessDenied
		}
		r.Repos.Study().AddPermission(queryPerm)
		study, err := r.Repos.Study().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	case "User":
		var user *repo.UserPermit
		viewerId, _ := viewer.ID()
		if args.Id == viewerId {
			user = viewer
		} else {
			queryPerm, err := r.Repos.Perm().GetQueryPermission(
				perm.ReadUser,
				viewer.Roles()...,
			)
			if err != nil {
				mylog.Log.WithError(err).Error("error retrieving query permission")
				return nil, repo.ErrAccessDenied
			}
			r.Repos.User().AddPermission(queryPerm)
			user, err = r.Repos.User().Get(args.Id)
			if err != nil {
				return nil, err
			}
		}
		return &nodeResolver{&userResolver{User: user, Repos: r.Repos}}, nil
	default:
		return nil, errors.New("invalid id")
	}
}

func (r *RootResolver) Nodes(ctx context.Context, args struct {
	Ids []string
}) ([]*nodeResolver, error) {
	nodes := make([]*nodeResolver, len(args.Ids))
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		mylog.Log.Error("viewer not found")
	}
	for i, id := range args.Ids {
		parsedId, err := oid.Parse(id)
		if err != nil {
			return nil, err
		}
		switch parsedId.Type() {
		case "User":
			var user *repo.UserPermit
			viewerId, _ := viewer.ID()
			if id == viewerId {
				user = viewer
			} else {
				queryPerm, err := r.Repos.Perm().GetQueryPermission(
					perm.ReadUser,
					viewer.Roles()...,
				)
				if err != nil {
					mylog.Log.WithError(err).Error("error retrieving query permission")
					return nil, repo.ErrAccessDenied
				}
				r.Repos.User().AddPermission(queryPerm)
				user, err = r.Repos.User().Get(id)
				if err != nil {
					return nil, err
				}
			}
			nodes[i] = &nodeResolver{&userResolver{User: user, Repos: r.Repos}}
		default:
			return nil, errors.New("invalid id")
		}
	}
	return nodes, nil
}

func (r *RootResolver) Study(
	ctx context.Context,
	args struct {
		Name  string
		Owner string
	},
) (*studyResolver, error) {
	var study *repo.StudyPermit
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.ReadStudy,
		viewer.Roles()...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.Study().AddPermission(queryPerm)
	study, err = r.Repos.Study().GetByUserAndName(args.Owner, args.Name)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
func (r *RootResolver) User(ctx context.Context, args struct {
	Login string
}) (*userResolver, error) {
	var user *repo.UserPermit
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	login, _ := viewer.Login()
	if login == args.Login {
		user = viewer
	} else {
		queryPerm, err := r.Repos.Perm().GetQueryPermission(
			perm.ReadUser,
			viewer.Roles()...,
		)
		if err != nil {
			mylog.Log.WithError(err).Error("error retrieving query permission")
			return nil, repo.ErrAccessDenied
		}
		r.Repos.User().AddPermission(queryPerm)
		user, err = r.Repos.User().GetByLogin(args.Login)
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *RootResolver) Viewer(ctx context.Context) (*userResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	return &userResolver{User: viewer, Repos: r.Repos}, nil
}
