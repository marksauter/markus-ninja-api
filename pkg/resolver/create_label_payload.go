package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createLabelPayloadResolver struct {
	Conf    *myconf.Config
	Label   *repo.LabelPermit
	StudyID *mytype.OID
	Repos   *repo.Repos
}

func (r *createLabelPayloadResolver) LabelEdge() (*labelEdgeResolver, error) {
	return NewLabelEdgeResolver(r.Label, r.Repos, r.Conf)
}

func (r *createLabelPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
