package resolver

import (
	"context"
	"errors"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type RemoveLabelPayload = removeLabelPayloadResolver

type removeLabelPayloadResolver struct {
	LabelId     *mytype.OID
	LabelableId *mytype.OID
	Repos       *repo.Repos
}

func (r *removeLabelPayloadResolver) Label(
	ctx context.Context,
) (*labelResolver, error) {
	label, err := r.Repos.Label().Get(ctx, r.LabelId.String)
	if err != nil {
		return nil, err
	}
	return &labelResolver{Label: label, Repos: r.Repos}, nil
}

func (r *removeLabelPayloadResolver) Labelable(
	ctx context.Context,
) (*labelableResolver, error) {
	permit, err := r.Repos.GetLabelable(ctx, r.LabelableId)
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
