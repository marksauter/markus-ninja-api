package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createActivityPayloadResolver struct {
	Activity *repo.ActivityPermit
	Conf     *myconf.Config
	Repos    *repo.Repos
	StudyID  *mytype.OID
}

func (r *createActivityPayloadResolver) ActivityEdge() (*activityEdgeResolver, error) {
	return NewActivityEdgeResolver(r.Activity, r.Repos, r.Conf)
}

func (r *createActivityPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
