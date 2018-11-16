package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type deleteLessonCommentPayloadResolver struct {
	Conf            *myconf.Config
	LessonCommentID *mytype.OID
	LessonID        *mytype.OID
	Repos           *repo.Repos
}

func (r *deleteLessonCommentPayloadResolver) DeletedLessonCommentID() graphql.ID {
	return graphql.ID(r.LessonCommentID.String)
}

func (r *deleteLessonCommentPayloadResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lesson, err := r.Repos.Lesson().Get(ctx, r.LessonID.String)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}
