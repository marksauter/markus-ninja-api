package resolver

// import (
//   graphql "github.com/graph-gophers/graphql-go"
//   "github.com/marksauter/markus-ninja-api/pkg/repo"
// )
//
// type Email = emailResolver
//
// type emailResolver struct {
//   Email *repo.EmailPermit
//   Repos *repo.Repos
// }
//
// func (r *emailResolver) CreatedAt() (graphql.Time, error) {
//   t, err := r.Email.CreatedAt()
//   return graphql.Time{t}, err
// }
//
// func (r *emailResolver) ID() (graphql.ID, error) {
//   id, err := r.Email.ID()
//   return graphql.ID(id.String), err
// }
//
// func (r *emailResolver) Value() (string, error) {
//   return r.Email.Value()
// }
