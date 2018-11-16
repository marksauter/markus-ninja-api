package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteLabelPayloadResolver struct {
	Conf    *myconf.Config
	LabelID *mytype.OID
	StudyID *mytype.OID
	Repos   *repo.Repos
}

func (r *deleteLabelPayloadResolver) DeletedLabelID() graphql.ID {
	return graphql.ID(r.LabelID.String)
}

func (r *deleteLabelPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
