package resolver

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteLessonCommentPayload = deleteLessonCommentPayloadResolver

type deleteLessonCommentPayloadResolver struct {
	LessonCommentId *mytype.OID
	LessonId        *mytype.OID
	Repos           *repo.Repos
}

func (r *deleteLessonCommentPayloadResolver) DeletedLessonCommentId() graphql.ID {
	return graphql.ID(r.LessonCommentId.String)
}

func (r *deleteLessonCommentPayloadResolver) Lesson() (*lessonResolver, error) {
	lesson, err := r.Repos.Lesson().Get(r.LessonId.String)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}
