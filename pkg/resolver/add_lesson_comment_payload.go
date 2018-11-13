package resolver

import (
	"context"

	"github.com/marksauter/markus-ninja-api/pkg/myconf"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type addLessonCommentPayloadResolver struct {
	Conf          *myconf.Config
	LessonComment *repo.LessonCommentPermit
	Repos         *repo.Repos
}

func (r *addLessonCommentPayloadResolver) CommentEdge() (*lessonCommentEdgeResolver, error) {
	return NewLessonCommentEdgeResolver(r.LessonComment, r.Repos, r.Conf)
}

func (r *addLessonCommentPayloadResolver) Lesson(
	ctx context.Context,
) (*lessonResolver, error) {
	lessonID, err := r.LessonComment.LessonID()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(ctx, lessonID.String)
	if err != nil {
		return nil, err
	}

	return &lessonResolver{Lesson: lesson, Conf: r.Conf, Repos: r.Repos}, nil
}
