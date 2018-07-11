package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteLabelPayload = deleteLabelPayloadResolver

type deleteLabelPayloadResolver struct {
	LabelId *mytype.OID
	StudyId *mytype.OID
	Repos   *repo.Repos
}

func (r *deleteLabelPayloadResolver) DeletedLabelId() graphql.ID {
	return graphql.ID(r.LabelId.String)
}

func (r *deleteLabelPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyId.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Repos: r.Repos}, nil
}
