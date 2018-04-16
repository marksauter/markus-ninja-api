package resolver

import "github.com/marksauter/markus-ninja-api/pkg/repo"

type RootResolver struct {
	Repos *repo.Repos
}
