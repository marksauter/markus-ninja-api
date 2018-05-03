package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/mysmtp"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type RootResolver struct {
	MailSvc mysmtp.Mailer
	Repos   *repo.Repos
	Svcs    *data.Services
}
