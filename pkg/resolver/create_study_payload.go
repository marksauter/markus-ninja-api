package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createStudyPayloadResolver struct {
	Conf   *myconf.Config
	Repos  *repo.Repos
	Study  *repo.StudyPermit
	UserID *mytype.OID
}

func (r *createStudyPayloadResolver) StudyEdge() (*studyEdgeResolver, error) {
	return NewStudyEdgeResolver(r.Study, r.Repos, r.Conf)
}

func (r *createStudyPayloadResolver) User(ctx context.Context) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.UserID.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Conf: r.Conf, Repos: r.Repos}, nil
}
