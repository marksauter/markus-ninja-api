package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type Email = emailResolver

type emailResolver struct {
	Email *repo.EmailPermit
	Repos *repo.Repos
}

func (r *emailResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.Email.CreatedAt()
	return graphql.Time{t}, err
}

func (r *emailResolver) ID() (graphql.ID, error) {
	id, err := r.Email.ID()
	return graphql.ID(id.String), err
}

func (r *emailResolver) IsVerified() (bool, error) {
	return r.Email.IsVerified()
}

func (r *emailResolver) Public() (bool, error) {
	return r.Email.Public()
}

func (r *emailResolver) Type() (string, error) {
	return r.Email.Type()
}

func (r *emailResolver) User(ctx context.Context) (*userResolver, error) {
	userId, err := r.Email.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *emailResolver) Value() (string, error) {
	return r.Email.Value()
}

func (r *emailResolver) VerifiedAt() (*graphql.Time, error) {
	t, err := r.Email.VerifiedAt()
	if err != nil {
		return nil, err
	}
	if t != nil {
		gqlTime := graphql.Time{*t}
		return &gqlTime, err
	}
	return nil, nil
}
