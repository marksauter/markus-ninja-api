package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddLabelPayload = addLabelPayloadResolver

type addLabelPayloadResolver struct {
	LabelID     *mytype.OID
	LabelableID *mytype.OID
	Repos       *repo.Repos
}

func (r *addLabelPayloadResolver) LabelEdge(
	ctx context.Context,
) (*labelEdgeResolver, error) {
	labelPermit, err := r.Repos.Label().Get(ctx, r.LabelID.String)
	if err != nil {
		return nil, err
	}
	return NewLabelEdgeResolver(labelPermit, r.Repos)
}

func (r *addLabelPayloadResolver) Labelable(
	ctx context.Context,
) (*labelableResolver, error) {
	permit, err := r.Repos.GetLabelable(ctx, r.LabelableID)
	if err != nil {
		return nil, err
	}
	resolver, err := nodePermitToResolver(permit, r.Repos)
	if err != nil {
		return nil, err
	}
	labelable, ok := resolver.(labelable)
	if !ok {
		return nil, errors.New("cannot convert resolver to labelable")
	}
	return &labelableResolver{labelable}, nil
}
