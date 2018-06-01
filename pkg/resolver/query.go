package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
)

func (r *RootResolver) Asset(
	ctx context.Context,
	args struct {
		Owner string
		Study string
		Name  string
	},
) (*userAssetResolver, error) {
	userAsset, err := r.Repos.UserAsset().GetByUserStudyAndName(
		args.Owner,
		args.Study,
		args.Name,
	)
	if err != nil {
		return nil, err
	}
	return &userAssetResolver{UserAsset: userAsset, Repos: r.Repos}, nil
}

func (r *RootResolver) Node(
	ctx context.Context,
	args struct{ Id string },
) (*nodeResolver, error) {
	parsedId, err := oid.Parse(args.Id)
	if err != nil {
		return nil, err
	}
	switch parsedId.Type {
	case "Lesson":
		lesson, err := r.Repos.Lesson().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&lessonResolver{Lesson: lesson, Repos: r.Repos}}, nil
	case "Study":
		study, err := r.Repos.Study().Get(args.Id)
		if err != nil {
			return nil, err
		}
		return &nodeResolver{&studyResolver{Study: study, Repos: r.Repos}}, nil
	case "User":
		user, err := r.Repos.User().Get(args.Id)
		if err != nil {
			return nil, err
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
	for i, id := range args.Ids {
		parsedId, err := oid.Parse(id)
		if err != nil {
			return nil, err
		}
		switch parsedId.Type {
		case "User":
			user, err := r.Repos.User().Get(id)
			if err != nil {
				return nil, err
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
	study, err := r.Repos.Study().GetByUserLoginAndName(args.Owner, args.Name)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
func (r *RootResolver) User(ctx context.Context, args struct {
	Login string
}) (*userResolver, error) {
	user, err := r.Repos.User().GetByLogin(args.Login)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *RootResolver) Viewer(ctx context.Context) (*userResolver, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	user, err := r.Repos.User().Get(viewer.Id.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}
