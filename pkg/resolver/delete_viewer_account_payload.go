package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteViewerAccountPayloadResolver struct {
	Conf     *myconf.Config
	Repos    *repo.Repos
	ViewerID *mytype.OID
}

func (r *deleteViewerAccountPayloadResolver) DeletedViewerID(
	ctx context.Context,
) graphql.ID {
	return graphql.ID(r.ViewerID.String)
}
