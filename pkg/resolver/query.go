package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

func (r *RootResolver) Node(
	ctx context.Context,
	args struct{ Id string },
) (*nodeResolver, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		mylog.Log.Error("viewer not found")
	}
	parsedId, err := oid.Parse(args.Id)
	if err != nil {
		return nil, err
	}
	switch parsedId.Type {
	case "Lesson":
		_, err := r.Repos.Lesson().AddPermission(perm.ReadLesson)
		if err != nil {
			return nil, err
		}
		lesson, err := r.Repos.Lesson().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&lessonResolver{Lesson: lesson, Repos: r.Repos}}, nil
	case "Study":
		_, err := r.Repos.Study().AddPermission(perm.ReadStudy)
		if err != nil {
			return nil, err
		}
		study, err := r.Repos.Study().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	case "User":
		var user *repo.UserPermit
		viewerId, _ := viewer.ID()
		if args.Id == viewerId.String {
			user = viewer
		} else {
			_, err := r.Repos.User().AddPermission(perm.ReadUser)
			if err != nil {
				return nil, err
			}
			user, err = r.Repos.User().Get(args.Id)
			if err != nil {
				return nil, err
			}
			err = user.ViewerCanAdmin(ctx)
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
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		mylog.Log.Error("viewer not found")
	}
	for i, id := range args.Ids {
		parsedId, err := oid.Parse(id)
		if err != nil {
			return nil, err
		}
		switch parsedId.Type {
		case "User":
			var user *repo.UserPermit
			viewerId, _ := viewer.ID()
			if id == viewerId.String {
				user = viewer
			} else {
				_, err := r.Repos.User().AddPermission(perm.ReadUser)
				if err != nil {
					return nil, err
				}
				user, err = r.Repos.User().Get(id)
				if err != nil {
					return nil, err
				}
				err = user.ViewerCanAdmin(ctx)
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
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	_, err := r.Repos.Study().AddPermission(perm.ReadStudy, viewer.Roles()...)
	if err != nil {
		return nil, err
	}
	study, err = r.Repos.Study().GetByUserLoginAndName(args.Owner, args.Name)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
func (r *RootResolver) User(ctx context.Context, args struct {
	Login string
}) (*userResolver, error) {
	var user *repo.UserPermit
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	login, _ := viewer.Login()
	if login == args.Login {
		user = viewer
	} else {
		_, err := r.Repos.User().AddPermission(perm.ReadUser)
		if err != nil {
			return nil, err
		}
		user, err = r.Repos.User().GetByLogin(args.Login)
		if err != nil {
			return nil, err
		}
		err = user.ViewerCanAdmin(ctx)
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *RootResolver) Viewer(ctx context.Context) (*userResolver, error) {
	viewer, ok := repo.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	err := viewer.ViewerCanAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: viewer, Repos: r.Repos}, nil
}
