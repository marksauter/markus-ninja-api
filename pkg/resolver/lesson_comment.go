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

func (r *lessonCommentResolver) Author() (*userResolver, error) {
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return nil, err
	}
	user, err := r.Repos.User().Get(userId)
	if err != nil {
		return nil, err
	}
	return &userResolver{User: user, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) Body() (string, error) {
	return r.Body()
}

func (r *lessonCommentResolver) BodyHTML() (mygql.HTML, error) {
	body, err := r.Body()
	if err != nil {
		return "", err
	}
	h := mygql.HTML(fmt.Sprintf("<div>%v</div>", body))
	return h, nil
}

func (r *lessonCommentResolver) CreatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.CreatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) ID() (graphql.ID, error) {
	id, err := r.LessonComment.ID()
	return graphql.ID(id), err
}

func (r *lessonCommentResolver) LessonId() (string, error) {
	return r.LessonComment.LessonId()
}

func (r *lessonCommentResolver) PublishedAt() (graphql.Time, error) {
	t, err := r.LessonComment.PublishedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) ResourcePath() (mygql.URI, error) {
	var uri mygql.URI
	study, err := r.Study()
	if err != nil {
		return uri, err
	}
	studyResourcePath, err := study.ResourcePath()
	if err != nil {
		return uri, err
	}
	lessonId, err := r.LessonId()
	if err != nil {
		return uri, err
	}
	lesson, err := r.Repos.Lesson().Get(lessonId)
	if err != nil {
		return uri, err
	}
	lessonNumber, err := lesson.Number()
	if err != nil {
		return uri, err
	}
	createdAt, err := r.CreatedAt()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s/lesson/%s#%s", studyResourcePath, lessonNumber, createdAt.Unix()))
	return uri, nil
}

func (r *lessonCommentResolver) Study() (*studyResolver, error) {
	studyId, err := r.LessonComment.StudyId()
	if err != nil {
		return nil, err
	}
	study, err := r.Repos.Study().Get(studyId)
	if err != nil {
		return nil, err
	}
	return &studyResolver{Study: study, Repos: r.Repos}, nil
}

func (r *lessonCommentResolver) UpdatedAt() (graphql.Time, error) {
	t, err := r.LessonComment.UpdatedAt()
	return graphql.Time{t}, err
}

func (r *lessonCommentResolver) URL() (mygql.URI, error) {
	var uri mygql.URI
	resourcePath, err := r.ResourcePath()
	if err != nil {
		return uri, err
	}
	uri = mygql.URI(fmt.Sprintf("%s%s", clientURL, resourcePath))
	return uri, nil
}

func (r *lessonCommentResolver) ViewerDidAuthor(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}

func (r *lessonCommentResolver) ViewerCanUpdate(ctx context.Context) (bool, error) {
	viewer, ok := myctx.User.FromContext(ctx)
	if !ok {
		return false, errors.New("viewer not found")
	}
	viewerId, _ := viewer.ID()
	userId, err := r.LessonComment.UserId()
	if err != nil {
		return false, err
	}

	return viewerId == userId, nil
}
