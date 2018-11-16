package resolver

import (
	"context"

	graphql "github.com/marksauter/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
)

type comment interface {
	Author(ctx context.Context) (*userResolver, error)
	Body() (string, error)
	BodyHTML(ctx context.Context) (mygql.HTML, error)
	ID() (graphql.ID, error)
	ViewerDidAuthor(ctx context.Context) (bool, error)
}

type commentResolver struct {
	comment
}

func (r *commentResolver) ToLessonComment() (*lessonCommentResolver, bool) {
	resolver, ok := r.comment.(*lessonCommentResolver)
	return resolver, ok
}
