package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteActivityPayloadResolver struct {
	ActivityID *mytype.OID
	Conf       *myconf.Config
	StudyID    *mytype.OID
	Repos      *repo.Repos
}

func (r *deleteActivityPayloadResolver) DeletedActivityID(
	ctx context.Context,
) graphql.ID {
	return graphql.ID(r.ActivityID.String)
}

func (r *deleteActivityPayloadResolver) Study(
	ctx context.Context,
) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
