package resolver

import (
	"context"
	"errors"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type removeLabelPayloadResolver struct {
	Conf        *myconf.Config
	LabelID     *mytype.OID
	LabelableID *mytype.OID
	Repos       *repo.Repos
}

func (r *removeLabelPayloadResolver) Labelable(
	ctx context.Context,
) (*labelableResolver, error) {
	permit, err := r.Repos.GetLabelable(ctx, r.LabelableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos, r.Conf)
	if err != nil {
		return nil, err
	}
	labelable, ok := resolver.(labelable)
	if !ok {
		return nil, errors.New("cannot convert resolver to labelable")
	}

	return &labelableResolver{labelable}, nil
}

func (r *removeLabelPayloadResolver) RemovedLabelID() graphql.ID {
	return graphql.ID(r.LabelID.String)
}
