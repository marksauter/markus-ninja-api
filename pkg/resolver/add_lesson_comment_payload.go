package resolver

import (
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type AddLessonCommentPayload = addLessonCommentPayloadResolver

type addLessonCommentPayloadResolver struct {
	LessonComment *repo.LessonCommentPermit
	Repos         *repo.Repos
}

func (r *addLessonCommentPayloadResolver) CommentEdge() (*lessonCommentEdgeResolver, error) {
	return NewLessonCommentEdgeResolver(r.LessonComment, r.Repos)
}

func (r *addLessonCommentPayloadResolver) Lesson() (*lessonResolver, error) {
	id, err := r.LessonComment.ID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(id.String)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}
