package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type DeleteLessonCommentPayload = deleteLessonCommentPayloadResolver

type deleteLessonCommentPayloadResolver struct {
	LessonCommentId       *mytype.OID
	LessonTimelineEventId *mytype.OID
	LessonId              *mytype.OID
	Repos                 *repo.Repos
}

func (r *deleteLessonCommentPayloadResolver) DeletedLessonCommentId() graphql.ID {
	return graphql.ID(r.LessonCommentId.String)
}

func (r *deleteLessonCommentPayloadResolver) DeletedLessonTimelineEventId() graphql.ID {
	return graphql.ID(r.LessonTimelineEventId.String)
}

func (r *deleteLessonCommentPayloadResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lesson, err := r.Repos.Lesson().Get(ctx, r.LessonId.String)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}
