package resolver

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/mygql"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type LessonComment = lessonCommentResolver

type lessonCommentResolver struct {
	LessonComment *repo.LessonCommentPermit
	Repos         *repo.Repos
}

func (r *lessonCommentResolver) Author(ctx context.Context) (*userResolver, error) {
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(ctx, userId.String)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) Body() (string, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	return body.String, nil
}

func (r *lessonCommentResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	return mygql.HTML(body.ToHTML()), nil
}

func (r *lessonCommentResolver) BodyText() (string, error) {
	body, err := r.LessonComment.Body()
	if err != nil {
		return "", err
	}
	return body.ToText(), nil
}

func (r *lessonCommentResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) ID() (graphql.ID, error) {
	id, err := r.LessonComment.ID()
	return graphql.ID(id.String), err
}

func (r *lessonCommentResolver) Lesson(ctx context.Context) (*lessonResolver, error) {
	lessonId, err := r.LessonComment.LessonId()
	if err != nil {
		return nil, err
	}
	lesson, err := r.Repos.Lesson().Get(ctx, lessonId.String)
	if err != nil {
		return nil, err
	}
	return &lessonResolver{Lesson: lesson, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) PublishedAt() (*graphql.Time, error) {
	t, err := r.LessonComment.PublishedAt()
	if err != nil {
		return nil, err
	}
	return &graphql.Time{t}, nil
}

func (r *lessonCommentResolver) ResourcePath(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	lesson, err := r.Lesson(ctx)
	if err != nil {
		return uri, err
	}
	lessonPath, err := lesson.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	createdAt, err := r.LessonComment.CreatedAt()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf(
		"%s/%d#lesson-comment%d",
		string(lessonPath),
		createdAt.Unix(),
	))
	return uri, nil
}

func (r *lessonCommentResolver) Study(ctx context.Context) (*studyResolver, error) {
	studyId, err := r.LessonComment.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(ctx, studyId.String)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) URL(
	ctx context.Context,
) (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath(ctx)
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonCommentResolver) ViewerCanDelete(ctx context.Context) bool {
	lessonComment := r.LessonComment.Get()
	return r.Repos.LessonComment().ViewerCanDelete(ctx, lessonComment)
}

func (r *lessonCommentResolver) ViewerCanUpdate(ctx context.Context) bool {
	lessonComment := r.LessonComment.Get()
	return r.Repos.LessonComment().ViewerCanUpdate(ctx, lessonComment)
}

func (r *lessonCommentResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.UserFromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return false, err
	}

	return viewer.Id.String == userId.String, nil
}
