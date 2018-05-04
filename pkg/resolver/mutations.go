package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type CreateUserInput struct {
	Email    string
	Login    string
	Password string
}

func (r *RootResolver) CreateUser(
	ctx context.Context,
	args struct{ Input CreateUserInput },
) (*userResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.CreateUser,
		viewer.Roles()...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to retrieve query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.User().AddPermission(*queryPerm)

	var user data.UserModel

	password, err := passwd.New(args.Input.Password)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create password")
		return nil, err
	}
	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		mylog.Log.WithError(err).Error("password failed strength check")
		return nil, err
	}

	user.Login.Set(args.Input.Login)
	user.Password.Set(password.Hash())
	user.PrimaryEmail.Set(args.Input.Email)

	userPermit, err := r.Repos.User().Create(&user)
	if err != nil {
		return nil, err
	}
	uResolver := &userResolver{userPermit}

	avt := &data.EmailVerificationTokenModel{}
	avt.UserId.Set(user.Id.String)

	err = r.Svcs.AVT.Create(avt)
	if err != nil {
		return uResolver, err
	}

	err = r.Svcs.Mail.SendEmailVerificationMail(
		user.PrimaryEmail.String,
		user.Login.String,
		avt.Token.String,
	)
	if err != nil {
		return uResolver, err
	}

	return uResolver, nil
}

type DeleteUserInput struct {
	Id string
}

func (r *RootResolver) DeleteUser(
	ctx context.Context,
	args struct{ Input DeleteUserInput },
) (*string, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	roles := viewer.Roles()
	if viewerId == args.Input.Id {
		roles = append(roles, "SELF")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.DeleteUser,
		roles...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.User().AddPermission(*queryPerm)

	id, err := oid.Parse(args.Input.Id)
	if err != nil {
		return nil, err
	}

	err = r.Repos.User().Delete(id.String())
	if err != nil {
		return nil, err
	}

	return &args.Input.Id, nil
}

type UpdateUserInput struct {
	Bio   *string
	Id    string
	Login *string
	Name  *string
}

func (r *RootResolver) UpdateUser(
	ctx context.Context,
	args struct{ Input UpdateUserInput },
) (*userResolver, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return nil, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	roles := viewer.Roles()
	if viewerId == args.Input.Id {
		roles = append(roles, "SELF")
	}
	queryPerm, err := r.Repos.Perm().GetQueryPermission(
		perm.UpdateUser,
		roles...,
	)
	if err != nil {
		mylog.Log.WithError(err).Error("error retrieving query permission")
		return nil, repo.ErrAccessDenied
	}
	r.Repos.User().AddPermission(*queryPerm)

	var user data.UserModel

	id, err := oid.Parse(args.Input.Id)
	if err != nil {
		return nil, err
	}
	user.Id.Set(id.String())

	if args.Input.Bio != nil {
		user.Bio.Set(args.Input.Bio)
	}
	if args.Input.Login != nil {
		user.Login.Set(args.Input.Login)
	}
	if args.Input.Name != nil {
		user.Name.Set(args.Input.Name)
	}

	userPermit, err := r.Repos.User().Update(&user)
	if err != nil {
		return nil, err
	}
	return &userResolver{userPermit}, nil
}
