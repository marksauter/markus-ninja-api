package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteLessonPayloadResolver struct {
	Conf     *myconf.Config
	LessonID *mytype.OID
	StudyID  *mytype.OID
	Repos    *repo.Repos
}

func (r *deleteLessonPayloadResolver) DeletedLessonID() graphql.ID {
	return graphql.ID(r.LessonID.String)
}

func (r *deleteLessonPayloadResolver) Study(ctx context.Context) (*studyResolver, error) {
	study, err := r.Repos.Study().Get(ctx, r.StudyID.String)
	if err != nil {
		return nil, err
	}

	return &studyResolver{Study: study, Conf: r.Conf, Repos: r.Repos}, nil
}
