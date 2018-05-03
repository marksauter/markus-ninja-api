package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

type RootResolver struct {
	Repos *repo.Repos
	Svcs  *service.Services
}
