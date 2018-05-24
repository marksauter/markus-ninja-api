package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type EVT = evtResolver

type evtResolver struct {
	EVT   *repo.EVTPermit
	Repos *repo.Repos
}

func (r *evtResolver) ExpiresAt() (graphql.Time, error) {
	t, err := r.EVT.ExpiresAt()
	return graphql.Time{t}, err
}

func (r *evtResolver) IssuedAt() (graphql.Time, error) {
	t, err := r.EVT.IssuedAt()
	return graphql.Time{t}, err
}

func (r *evtResolver) VerifiedAt() (*graphql.Time, error) {
	t, err := r.EVT.VerifiedAt()
	return &graphql.Time{t}, err
}