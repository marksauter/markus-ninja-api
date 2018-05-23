package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
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
) (*emailResolver, error) {
	id, err := r.UserEmail.EmailId()
	if err != nil {
		return nil, err
	}
	email, err := r.Repos.Email().Get(id)
	if err != nil {
		return nil, err
	}
	return &emailResolver{Email: email, Repos: r.Repos}, nil
}

func (r *userEmailResolver) Type() (string, error) {
	return r.UserEmail.Type()
}

func (r *userEmailResolver) User(ctx context.Context) (*userResolver, error) {
	userId, err := r.UserEmail.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *userEmailResolver) VerifiedAt() (graphql.Time, error) {
	t, err := r.UserEmail.VerifiedAt()
	return graphql.Time{t}, err
}
