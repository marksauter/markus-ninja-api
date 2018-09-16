package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type createStudyPayloadResolver struct {
	Study  *repo.StudyPermit
	UserID *mytype.OID
	Repos  *repo.Repos
}

func (r *createStudyPayloadResolver) StudyEdge() (*studyEdgeResolver, error) {
	return NewStudyEdgeResolver(r.Study, r.Repos)
}

func (r *createStudyPayloadResolver) User(ctx context.Context) (*userResolver, error) {
	user, err := r.Repos.User().Get(ctx, r.UserID.String)
	if err != nil {
		return nil, err
	}

	return &userResolver{User: user, Repos: r.Repos}, nil
}
