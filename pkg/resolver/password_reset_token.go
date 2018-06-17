package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type PRT = prtResolver

type prtResolver struct {
	PRT   *repo.PRTPermit
	Repos *repo.Repos
}

func (r *prtResolver) ExpiresAt() (graphql.Time, error) {
	t, err := r.PRT.ExpiresAt()
	return graphql.Time{t}, err
}

func (r *prtResolver) IssuedAt() (graphql.Time, error) {
	t, err := r.PRT.IssuedAt()
	return graphql.Time{t}, err
}
