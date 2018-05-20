package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/perm"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type UserEmail = userEmailResolver

type userEmailResolver struct {
	UserEmail *repo.UserEmailPermit
	Repos     *repo.Repos
}

func (r *userEmailResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.UserEmail.CreatedAt()
	return graphql.Time{t}, err
}

func (r *userEmailResolver) Email(
	ctx context.Context,
	args struct{ Number int32 },
) (*lessonResolver, error) {
	_, err := r.Repos.Email().AddPermission(perm.ReadEmail)
	if err != nil {
		return nil, err
	}
	id, err := r.UserEmail.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Email().GetByPK(id)
	if err != nil {
		return nil, err
	}
	return &emailResolver{Email: email, Repos: r.Repos}, nil
}

func (r *userEmailResolver) ID() (graphql.ID, error) {
	id, err := r.UserEmail.ID()
	return graphql.ID(id), err
}

func (r *userEmailResolver) Type() (string, error) {
	return r.UserEmail.Type()
}

func (r *userEmailResolver) User(ctx context.Context) (*userResolver, error) {
	userId, err := r.UserEmail.UserId()
	if err != nil {
		return nil, err
	}
	_, err = r.Repos.User().AddPermission(perm.ReadUser)
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	err = user.ViewerCanAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *userEmailResolver) VerfiedAt() (graphql.Time, error) {
	t, err := r.UserEmail.VerfiedAt()
	return graphql.Time{t}, err
}
