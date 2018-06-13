package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
	"github.com/marksauter/markus-ninja-api/pkg/service"
)

var clientURL = "http://localhost:3000"

type RootResolver struct {
	Repos *repo.Repos
	Svcs  *service.Services
}
